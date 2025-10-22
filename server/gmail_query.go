package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
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
	rec := TokenRecord{
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

func gmailClientForUser(r *gin.Context, accountId string) (*gmail.Service, error) {
	rec, err := loadGmailTokenRecord(r, accountId)
	if err != nil {
		return nil, err
	}
	if rec == nil || rec.RefreshToken == "" {
		return nil, errors.New("Invalid token")
	}

	// Seed token from DB
	baseTok := &oauth2.Token{
		AccessToken:  rec.AccessToken,
		RefreshToken: rec.RefreshToken,
		TokenType:    rec.TokenType,
		Expiry:       rec.Expiry,
	}

	// Base source that knows how to refresh via Google
	baseSrc := oauthConfig.TokenSource(r, baseTok)

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
	svc, err := gmail.NewService(r, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func ListInbox(r *gin.Context) {
	claimsI := r.Value("claims")
	if claimsI == nil {
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Not authorized"})
		return
	}
	claims := claimsI.(desktopClaims)
	svc, err := gmailClientForUser(r, claims.Subject)
	if err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
		return
	}
	res, err := ListInboxFor(r, svc, 100)
	if err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get query gmail"})
		return
	}
	r.JSON(http.StatusOK, res)
}

// List the most recent N messages in INBOX (with pagination)
func ListInboxFor(ctx context.Context, svc *gmail.Service, max int64) ([]*gmail.Message, error) {
	var out []*gmail.Message
	pageToken := ""
	remaining := max

	for remaining > 0 {
		pageSize := min(remaining, 100)

		listCall := svc.Users.Messages.
			List("me").
			LabelIds("INBOX").
			IncludeSpamTrash(false).
			MaxResults(pageSize).
			Fields(googleapi.Field("messages(id,threadId),nextPageToken")) // partial response

		if pageToken != "" {
			listCall = listCall.PageToken(pageToken)
		}

		listRes, err := listCall.Do()
		if err != nil {
			return nil, err
		}
		if len(listRes.Messages) == 0 {
			break
		}

		// Fetch lightweight metadata for each message (Subject/From/Date)
		for _, m := range listRes.Messages {
			msg, err := svc.Users.Messages.
				Get("me", m.Id).
				Format("metadata").
				MetadataHeaders("Subject", "From", "Date", "To").
				Fields(googleapi.Field("id,threadId,payload/headers,snippet,internalDate")).
				Do()
			if err != nil {
				return nil, err
			}
			out = append(out, msg)
			remaining--
			if remaining == 0 {
				break
			}
		}

		if listRes.NextPageToken == "" || remaining == 0 {
			break
		}
		pageToken = listRes.NextPageToken
	}
	return out, nil
}
