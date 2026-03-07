package oauth_basic

import (
	"fromkeith/my-desktop-server/shared/globals"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func CreateAccount(r *gin.Context, accountId string) error {
	_, err := globals.Db().Exec(r, `
		INSERT INTO UserAccounts (
			accountId,
			createdAt
		) VALUES ($1, $2)
		`, accountId, time.Now().UTC())
	return err
}

func SaveSession(r *gin.Context, session map[string]string) error {
	claimedId := r.GetString("accountId")
	_, err := globals.Db().Exec(r, `
		INSERT INTO OauthInitSession(
			state,
			claimedId,
			codeVerifier,
			postAuthReturn,
			createdAt,
			expiresAt
		) VALUES (
		$1,$2,
		$3,$4,
		$5,$6
		)
		`,
		session["state"],
		claimedId,
		session["code_verifier"],
		session["post_auth_return"],
		time.Now().UTC(),
		time.Now().UTC().Add(time.Minute*30),
	)
	return err
}

// TODO: this is lazy and dumb to use a map
func MustLoadSession(r *gin.Context, source_state string) map[string]string {

	row := globals.Db().QueryRow(r, `
		SELECT
			state,
			claimedId,
			codeVerifier,
			postAuthReturn,
			createdAt,
			expiresAt
		FROM OauthInitSession
		WHERE state = $1
		`,
		source_state)

	var state, code_verifier, post_auth_return, existingId string
	var expires_at, created_at time.Time
	err := row.Scan(
		&state,
		&existingId,
		&code_verifier,
		&post_auth_return,
		&created_at,
		&expires_at,
	)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to load OauthInitSession")
		return make(map[string]string)
	}
	return map[string]string{
		"state":            state,
		"claimed_id":       existingId,
		"code_verifier":    code_verifier,
		"post_auth_return": post_auth_return,
		"expires_at":       expires_at.Format(time.RFC3339Nano),
		"created_at":       created_at.Format(time.RFC3339Nano),
	}
}
