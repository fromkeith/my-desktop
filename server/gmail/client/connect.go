package client

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fromkeith/my-desktop-server/auth"
	"fromkeith/my-desktop-server/globals"
	oauth_basic "fromkeith/my-desktop-server/oauth"
	"fromkeith/my-desktop-server/utils"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/people/v1"
)

var (
	oauthConfig  *oauth2.Config
	oidcProvider *oidc.Provider
	oidcVerifier *oidc.IDTokenVerifier
)

func SetupGoogle() {
	creds := os.Getenv("GOOGLE_CREDENTIALS")
	var err error
	oauthConfig, err = google.ConfigFromJSON([]byte(creds), gmail.GmailReadonlyScope,
		"openid",
		"email", "profile",
		people.ContactsReadonlyScope,
		people.ContactsOtherReadonlyScope,
		people.DirectoryReadonlyScope,
	)
	if err != nil {
		log.Fatalf("Unable to parse client secret to config: %v", err)
	}
	oidcProvider, err = oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		panic(err)
	}
	oidcVerifier = oidcProvider.Verifier(&oidc.Config{ClientID: oauthConfig.ClientID})

}

func HandleAuthStart(r *gin.Context) {
	state := utils.RandB64(32)
	codeVerifier := utils.RandB64(64)
	nonce := utils.RandB64(32)

	err := oauth_basic.SaveSession(r, map[string]string{
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

	codeChallenge := base64.RawURLEncoding.EncodeToString(utils.Sha256Bytes(codeVerifier))
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

func loadGmailTokenRecord(r *gin.Context, accountId string) (*oauth_basic.TokenRecord, error) {
	return oauth_basic.LoadTokenRecord(r, accountId, "google")
}

func SaveGmailTokenRecord(r context.Context, accountId string, rec oauth_basic.TokenRecord) error {
	return oauth_basic.SaveTokenRecord(r, accountId, rec)
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

func HandleCallback(r *gin.Context) {
	code := r.Query("code")
	state := r.Query("state")
	sess := oauth_basic.MustLoadSession(r, state)
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
	assignAuth := false
	if existingAccountId == "" {
		assignAuth = true

		row := globals.Db().QueryRowContext(r, `
			SELECT u.account_id
			FROM user_oauth_accounts u
			WHERE u.user_id = ?
			`,
			claims.Sub,
		)
		// failed to get existing account... create a new one?
		if err := row.Scan(&existingAccountId); err != nil {

			if err == sql.ErrNoRows {
				existingAccountId = "acct_" + uuid.New().String()
				isNewUser = true
				err := oauth_basic.CreateAccount(r, existingAccountId)
				if err != nil {
					log.Println(err)
					r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to create account"})
					return
				}
			} else {
				log.Println(err)
				r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
				return
			}
		}
	}

	rec := oauth_basic.TokenRecord{
		UserId:       claims.Sub,
		Provider:     "google",
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken, // may be empty if you forgot access_type=offline or prompt=consent
		Expiry:       tok.Expiry,
		TokenType:    tok.TokenType,
		Scope:        "", // optional: persist actual granted scope string
	}
	if err := SaveGmailTokenRecord(r, existingAccountId, rec); err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to save token"})
		return
	}

	extraQuery := ""
	if assignAuth {
		claims := auth.DesktopClaims{}
		claims.Subject = existingAccountId
		tokenString, err := auth.CreateToken(claims)
		if err != nil {
			log.Println(err)
			r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "token failed to be created"})
			return
		}
		extraQuery = "?auth=" + tokenString

		if isNewUser {
			bkg := context.Background()
			client, _ := GmailClient(bkg, existingAccountId, true)
			go client.Bootstrap(bkg)
		}
	}
	// Done — redirect back to your app
	r.Redirect(http.StatusFound, os.Getenv("DOMAIN_URL")+sess["post_auth_return"]+extraQuery)
}
