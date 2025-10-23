package oauth_basic

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenRecord struct {
	UserId       string
	Provider     string // "google"
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	TokenType    string
	Scope        string
}

func CreateAccount(r *gin.Context, accountId string) error {
	_, err := globals.Db().ExecContext(r, `
		INSERT INTO user_accounts (
			account_id,
			created_at
		) VALUES (?, ?)
		`, accountId, time.Now().Format(time.RFC3339))
	return err
}

func SaveSession(r *gin.Context, session map[string]string) error {
	claimedId := r.GetString("accountId")
	_, err := globals.Db().ExecContext(r, `
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
func MustLoadSession(r *gin.Context, source_state string) map[string]string {

	row := globals.Db().QueryRowContext(r, `
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

func LoadTokenRecord(ctx context.Context, accountId, provider string) (*TokenRecord, error) {
	row := globals.Db().QueryRowContext(ctx, `
		SELECT o.user_id, access_token, refresh_token, expiry, token_type, scope
		FROM user_oauth_accounts u
		INNER JOIN oauth_token_record o
		WHERE u.account_id = ?
		AND o.provider = ?
		LIMIT 1
		`, accountId, provider)
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
	log.Println("saved expiry is", expiry)
	rec.Expiry, _ = time.Parse(time.RFC3339, expiry)
	return &rec, nil
}

func SaveTokenRecord(r context.Context, accountId string, rec TokenRecord) error {
	tx, err := globals.Db().BeginTx(r, nil)
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
