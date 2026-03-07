package oauth_basic

import (
	"context"
	"fromkeith/my-desktop-server/shared/globals"
	"time"

	"github.com/rs/zerolog/log"
)

type TokenRecord struct {
	UserId       string
	Provider     string // "google"
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
	TokenType    string
	Scope        string
	Version      int64
}

func LoadTokenRecord(ctx context.Context, accountId, provider string) (*TokenRecord, error) {
	row := globals.Db().QueryRow(ctx, `
		SELECT o.userId, accessToken, refreshToken, expiry, tokenType, scope, version
		FROM UserOauthAccounts u
		INNER JOIN OauthTokenRecord o ON u.userId = o.userId
		WHERE u.accountId = $1
		AND o.provider = $2
		LIMIT 1
		`, accountId, provider)
	rec := TokenRecord{
		Provider: "google",
	}
	err := row.Scan(
		&rec.UserId,
		&rec.AccessToken,
		&rec.RefreshToken,
		&rec.Expiry,
		&rec.TokenType,
		&rec.Scope,
		&rec.Version,
	)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func SaveTokenRecord(r context.Context, accountId string, rec TokenRecord) error {
	tx, err := globals.Db().Begin(r)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to beging transaction to save token record")
		return err
	}
	_, err = tx.Exec(r, `
		INSERT INTO OauthTokenRecord (
			userId,
		    provider,
		    accessToken,
		    refreshToken,
		    expiry,
		    tokenType,
		    scope,
			updatedAt
		) VALUES (
			$1, $2,
			$3, $4,
			$5, $6,
			$7, $8
		) ON CONFLICT (userId) DO UPDATE SET
		    accessToken = EXCLUDED.accessToken,
		    refreshToken = EXCLUDED.refreshToken,
		    expiry = EXCLUDED.expiry,
		    tokenType = EXCLUDED.tokenType,
		    scope = EXCLUDED.scope,
		    updatedAt = EXCLUDED.updatedAt,
			version = OauthTokenRecord.version + 1
		`,
		rec.UserId,
		rec.Provider,
		rec.AccessToken,
		rec.RefreshToken,
		rec.Expiry.UTC(),
		rec.TokenType,
		rec.Scope,
		time.Now().UTC(),
	)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to save OAuthTokenRecord")
		tx.Rollback(r)
		return err
	}
	_, err = tx.Exec(r, `
		INSERT INTO UserOauthAccounts (
			accountId,
			userId
		) VALUES ($1, $2)
		ON CONFLICT (accountId, userId) DO NOTHING
		`,
		accountId,
		rec.UserId,
	)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to save UserOAuthAccount")
		tx.Rollback(r)
		return err
	}
	return tx.Commit(r)
}
