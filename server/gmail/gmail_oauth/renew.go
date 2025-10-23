package gmail_oauth

import (
	"context"
	"errors"
	oauth_basic "fromkeith/my-desktop-server/oauth"
	"log"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type storingTokenSource struct {
	accountId string
	userId    string
	provider  string
	inner     oauth2.TokenSource
}

func (s *storingTokenSource) Token() (*oauth2.Token, error) {
	t, err := s.inner.Token()
	if err != nil {
		return nil, err
	}
	log.Println("Saving renewed", t)
	rec := oauth_basic.TokenRecord{
		UserId:       s.userId,
		Provider:     "google",
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken, // may be empty if you forgot access_type=offline or prompt=consent
		Expiry:       t.Expiry,
		TokenType:    t.TokenType,
		Scope:        "", // optional: persist actual granted scope string
	}

	_ = saveGmailTokenRecord(context.Background(), s.accountId, rec)
	return t, nil
}

func GmailClientForUser(ctx context.Context, accountId string, setToBackground bool) (*gmail.Service, string, error) {
	rec, err := oauth_basic.LoadTokenRecord(ctx, accountId, "google")
	if err != nil {
		return nil, "", err
	}
	log.Println("laoded token record", rec)
	if rec == nil || rec.RefreshToken == "" {
		return nil, "", errors.New("Invalid token")
	}

	// Seed token from DB
	baseTok := &oauth2.Token{
		AccessToken:  rec.AccessToken,
		RefreshToken: rec.RefreshToken,
		TokenType:    rec.TokenType,
		Expiry:       rec.Expiry,
	}

	bkg := ctx
	if setToBackground {
		bkg = context.Background()
	}
	// Base source that knows how to refresh via Google
	baseSrc := oauthConfig.TokenSource(bkg, baseTok)

	// ReuseTokenSource caches until near expiry; when it refreshes, we want to persist.
	reuse := oauth2.ReuseTokenSource(baseTok, baseSrc)

	// Wrap to save refreshed tokens
	ts := &storingTokenSource{
		accountId: accountId,
		userId:    rec.UserId,
		provider:  "google",
		inner:     reuse,
	}

	// Either gmail.New or gmail.NewService; both work. NewService lets you pass options.
	svc, err := gmail.NewService(bkg, option.WithTokenSource(ts))
	if err != nil {
		return nil, "", err
	}
	return svc, rec.UserId, nil
}
