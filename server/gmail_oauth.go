package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	oauthConfig  *oauth2.Config
	oidcProvider *oidc.Provider
	oidcVerifier *oidc.IDTokenVerifier
)

func setupGoogle() {
	creds := os.Getenv("GOOGLE_CREDENTIALS")
	var err error
	oauthConfig, err = google.ConfigFromJSON([]byte(creds), gmail.GmailReadonlyScope, "openid", "email", "profile")
	if err != nil {
		log.Fatalf("Unable to parse client secret to config: %v", err)
	}
	oidcProvider, err = oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		panic(err)
	}
	oidcVerifier = oidcProvider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID})

}

func randB64(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func sha256Bytes(src string) []byte {
	hash := sha256.New()
	hash.Write([]byte(src))
	return hash.Sum(nil)
}

func handleAuthStart(r *gin.Context) {
	state := randB64(32)
	codeVerifier := randB64(64)
	nonce := randB64(32)

	err := saveSession(r, map[string]string{
		"state":            state,
		"code_verifier":    codeVerifier,
		"post_auth_return": r.Query("return_to"),
		"nonce":            nonce,
	})
	if err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "failed to save"})
		return
	}

	codeChallenge := base64.RawURLEncoding.EncodeToString(sha256Bytes(codeVerifier))
	url := oauthConfig.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("include_granted_scopes", "true"),
	)
	r.Redirect(http.StatusFound, url)
}

type TokenRecord struct {
	UserId       string
	Provider     string // "google"
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	TokenType    string
	Scope        string
}

func saveSession(r *gin.Context, session map[string]string) error {
	claimedId := ""
	claimsI := r.Value("claims")
	if claimsI != nil {
		claims := claimsI.(*desktopClaims)
		claimedId = claims.Subject
	}
	_, err := db.ExecContext(r, `
		INSERT INTO oauth_init_session(
			state,
			claimed_id,
			code_verifier,
			post_auth_return,
			created_at,
			expires_at
		) VALUES (
		?,?,
		?,?,
		?,?
		)
		`,
		session["state"],
		claimedId,
		session["code_verifier"],
		session["post_auth_return"],
		time.Now().Format(time.RFC3339),
		time.Now().Add(time.Minute*30).Format(time.RFC3339),
	)
	return err
}

// TODO: this is lazy and dumb to use a map
func mustLoadSession(r *gin.Context, source_state string) map[string]string {

	row := db.QueryRowContext(r, `
		SELECT
			state,
			claimed_id,
			code_verifier,
			post_auth_return,
			created_at,
			expires_at
		FROM oauth_init_session
		WHERE state = ?
		`,
		source_state)

	var state, code_verifier, post_auth_return, existingId string
	var expires_at, created_at string
	err := row.Scan(
		&state,
		&existingId,
		&code_verifier,
		&post_auth_return,
		&created_at,
		&expires_at,
	)
	if err != nil {
		log.Println(err)
		return make(map[string]string)
	}
	return map[string]string{
		"state":            state,
		"claimed_id":       existingId,
		"code_verifier":    code_verifier,
		"post_auth_return": post_auth_return,
		"expires_at":       expires_at,
		"created_at":       created_at,
	}
}
func createAccount(r *gin.Context, accountId string) error {
	_, err := db.ExecContext(r, `
		INSERT INTO user_accounts (
			account_id,
			created_at
		) VALUES (?, ?)
		`, accountId, time.Now().Format(time.RFC3339))
	return err
}

func loadGmailTokenRecord(r *gin.Context, accountId string) (*TokenRecord, error) {
	row := db.QueryRowContext(r, `
		SELECT o.user_id, access_token, refresh_token, expiry, token_type, scope
		FROM user_oauth_accounts u
		INNER JOIN oauth_token_record o
		WHERE u.account_id = ?
		AND o.provider = 'google'
		LIMIT 1
		`, accountId)
	rec := TokenRecord{
		Provider: "google",
	}
	var expiry string
	err := row.Scan(
		&rec.UserId,
		&rec.AccessToken,
		&rec.RefreshToken,
		&expiry,
		&rec.TokenType,
		&rec.Scope,
	)
	if err != nil {
		return nil, err
	}
	rec.Expiry, _ = time.Parse(expiry, time.RFC3339)
	return &rec, nil
}

func saveGmailTokenRecord(r context.Context, accountId string, rec TokenRecord) error {
	tx, err := db.BeginTx(r, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = tx.Exec(`
		INSERT INTO oauth_token_record (
			user_id,
		    provider,
		    access_token,
		    refresh_token,
		    expiry,
		    token_type,
		    scope
		) VALUES (
			?, ?,
			?, ?,
			?, ?,
			?
		) ON CONFLICT (user_id) DO UPDATE SET
		    access_token = excluded.access_token,
		    refresh_token = excluded.refresh_token,
		    expiry = excluded.expiry,
		    token_type = excluded.token_type,
		    scope = excluded.scope
		`,
		rec.UserId,
		rec.Provider,
		rec.AccessToken,
		rec.RefreshToken,
		rec.Expiry.Format(time.RFC3339),
		rec.TokenType,
		rec.Scope,
	)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	tx.Exec(`
		INSERT INTO user_oauth_accounts (
			account_id,
			user_id
		) VALUES (?, ?)
		ON CONFLICT DO NOTHING
		`,
		accountId,
		rec.UserId,
	)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// https://developers.google.com/identity/openid-connect/openid-connect
type googleOidcClaims struct {
	Sub           string `json:"sub"` // stable Google user id
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
	Name          string `json:"name"`
	Hd            string `json:"hd"` // GSuite domain, if present
}

func handleCallback(r *gin.Context) {
	code := r.Query("code")
	state := r.Query("state")
	sess := mustLoadSession(r, state)
	if len(sess) == 0 {
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "state not found"})
		return
	}
	wantState := sess["state"]
	verifier := sess["code_verifier"]
	if state == "" || state != wantState {
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "state mismatch"})
		return
	}
	expires, _ := time.Parse(time.RFC3339, sess["expires_at"])
	if expires.Before(time.Now()) {
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "expired"})
		return
	}
	log.Println("exchange code: " + code)
	log.Println("verifier: " + verifier)
	tok, err := oauthConfig.Exchange(r, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "exchange failed"})
		return
	}
	log.Println("token")
	log.Println(tok)
	rawIdToken, _ := tok.Extra("id_token").(string)
	if rawIdToken == "" {
		log.Println("no id_token")
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing id_token"})
		return
	}

	idt, err := oidcVerifier.Verify(r, rawIdToken)
	if err != nil {
		print("oidc failed")
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "verify id_token: " + err.Error()})
		return
	}
	var claims googleOidcClaims
	if err := idt.Claims(&claims); err != nil {
		r.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Bad claims: " + err.Error()})
		return
	}

	existingAccountId := sess["claimed_id"]
	// new user
	isNewUser := false
	if existingAccountId == "" {
		existingAccountId = "acct_" + uuid.New().String()
		isNewUser = true
		err := createAccount(r, existingAccountId)
		if err != nil {
			log.Println(err)
			r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
			return
		}
	}

	rec := TokenRecord{
		UserId:       claims.Sub,
		Provider:     "google",
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken, // may be empty if you forgot access_type=offline or prompt=consent
		Expiry:       tok.Expiry,
		TokenType:    tok.TokenType,
		Scope:        "", // optional: persist actual granted scope string
	}
	if err := saveGmailTokenRecord(r, existingAccountId, rec); err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
		return
	}

	extraQuery := ""
	if isNewUser {
		claims := desktopClaims{}
		claims.Subject = existingAccountId
		tokenString, err := CreateToken(claims)
		if err != nil {
			log.Println(err)
			r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "token failed to be created"})
			return
		}
		extraQuery = "?auth=" + tokenString
	}

	// Done — redirect back to your app
	r.Redirect(http.StatusFound, os.Getenv("DOMAIN_URL")+sess["post_auth_return"]+extraQuery)
}
