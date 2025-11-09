package client

import (
	"context"
	"errors"
	"fmt"
	"fromkeith/my-desktop-server/globals"
	oauth_basic "fromkeith/my-desktop-server/oauth"
	"fromkeith/my-desktop-server/utils"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
)

type storingTokenSource struct {
	accountId string
	userId    string
	provider  string
	inner     oauth2.TokenSource
	ctx       context.Context
	last      *oauth2.Token
	mu        sync.RWMutex // protects lastTok
}

type googleClient struct {
	gmail     *gmail.Service
	people    *people.Service
	userId    string
	accountId string
}

func (s *storingTokenSource) Token() (*oauth2.Token, error) {
	// read lock
	s.mu.RLock()
	last := s.last
	var nearExpiry bool
	if last == nil {
		nearExpiry = true
	} else if time.Until(last.Expiry) <= time.Minute*5+time.Millisecond {
		nearExpiry = true
	}
	s.mu.RUnlock()
	if !nearExpiry {
		return last, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// it changed by the time we acquired the lock
	if s.last != last {
		return s.last, nil
	}

	// create a new context so we don't block postgres
	ctx, cancel := context.WithTimeout(s.ctx, 15*time.Second)
	defer cancel()

	conn, err := globals.Db().Acquire(ctx)
	if err != nil {
		log.Error().
			Ctx(s.ctx).
			Err(err).
			Msg("failed to acquire database connection to renew token")
		return nil, err
	}
	defer conn.Release()

	lockKey := utils.HashToInt64(fmt.Sprintf("%s:%s:%s", s.accountId, s.userId, s.provider))

	if err := globals.PostgresLock(ctx, conn.Conn(), lockKey, time.Second); err != nil {
		log.Error().
			Ctx(s.ctx).
			Err(err).
			Msg("failed to acquire lock in postgres to refresh token")
		return nil, err
	}
	defer globals.PostgresUnlock(context.Background(), conn.Conn(), lockKey)

	// check if a different service renewed the token while we waited
	// for the postgres lock
	rec, err := oauth_basic.LoadTokenRecord(ctx, s.accountId, s.provider)
	if err == nil && rec != nil {
		if (last == nil || rec.AccessToken != last.AccessToken) && time.Until(rec.Expiry) > time.Minute {
			s.last = &oauth2.Token{
				AccessToken:  rec.AccessToken,
				RefreshToken: rec.RefreshToken,
				TokenType:    rec.TokenType,
				Expiry:       rec.Expiry,
			}
			return s.last, nil
		}
	}

	// refresh the token from our refresher
	t, err := s.inner.Token()
	if err != nil {
		log.Error().
			Ctx(s.ctx).
			Err(err).
			Msg("failed to load token from token source")
		return nil, err
	}
	s.last = t
	log.Debug().
		Msg("Saving new token")
	rec = &oauth_basic.TokenRecord{
		UserId:       s.userId,
		Provider:     s.provider,
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken, // may be empty if you forgot access_type=offline or prompt=consent
		Expiry:       t.Expiry,
		TokenType:    t.TokenType,
		Scope:        "", // optional: persist actual granted scope string
	}

	err = SaveGmailTokenRecord(context.Background(), s.accountId, *rec)
	if err != nil {
		log.Error().
			Ctx(s.ctx).
			Err(err).
			Msg("failed to save renewed token record!")
		return nil, err
	}
	return t, nil
}

func GoogleClientFor(ctx context.Context, accountId string) (*googleClient, error) {
	if accountId == "" {
		return nil, errors.New("Invalid accountId")
	}
	rec, err := oauth_basic.LoadTokenRecord(ctx, accountId, "google")
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("failed to load gmail token record")
		return nil, err
	}
	if rec == nil || rec.RefreshToken == "" {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("token is invalid, can't load it")
		return nil, errors.New("Invalid token")
	}

	// Seed token from DB
	baseTok := &oauth2.Token{
		AccessToken:  rec.AccessToken,
		RefreshToken: rec.RefreshToken,
		TokenType:    rec.TokenType,
		Expiry:       rec.Expiry,
	}

	bkg := context.WithoutCancel(ctx)
	// Base source that knows how to refresh via Google
	baseSrc := oauthConfig.TokenSource(bkg, baseTok)
	// ReuseTokenSource caches until near expiry; when it refreshes, we want to persist.
	reuse := oauth2.ReuseTokenSourceWithExpiry(baseTok, baseSrc, time.Minute*5)

	// Wrap to save refreshed tokens
	ts := &storingTokenSource{
		accountId: accountId,
		userId:    rec.UserId,
		provider:  "google",
		inner:     reuse,
		ctx:       bkg,
		last:      baseTok,
	}

	// Either gmail.New or gmail.NewService; both work. NewService lets you pass options.
	svc, err := gmail.NewService(bkg, option.WithTokenSource(ts))
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("can't create gmail service")
		return nil, err
	}
	peopleSvc, err := people.NewService(bkg, option.WithTokenSource(ts))
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("failed to create people service")
		return nil, err
	}
	return &googleClient{
		gmail:     svc,
		people:    peopleSvc,
		userId:    rec.UserId,
		accountId: accountId,
	}, nil
}
