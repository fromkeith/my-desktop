package client

import (
	"context"
	"os"

	oauth_basic "fromkeith/my-desktop-server/shared/oauth"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/people/v1"
)

var (
	OauthConfig  *oauth2.Config
	oidcProvider *oidc.Provider
	OidcVerifier *oidc.IDTokenVerifier
)

func init() {
	creds := os.Getenv("GOOGLE_CREDENTIALS")
	var err error
	OauthConfig, err = google.ConfigFromJSON([]byte(creds),
		gmail.GmailModifyScope,
		gmail.GmailSendScope,
		gmail.GmailComposeScope,
		gmail.GmailLabelsScope,
		"openid",
		"email", "profile",
		people.ContactsReadonlyScope,
		people.ContactsOtherReadonlyScope,
		people.DirectoryReadonlyScope,
	)
	if err != nil {
		log.Fatal().
			Stack().
			Err(err).
			Msg("Unable to parse client secret to config")

	}
	oidcProvider, err = oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		panic(err)
	}
	OidcVerifier = oidcProvider.Verifier(&oidc.Config{ClientID: OauthConfig.ClientID})

}

func SaveGmailTokenRecord(r context.Context, accountId string, rec oauth_basic.TokenRecord) error {
	return oauth_basic.SaveTokenRecord(r, accountId, rec)
}
