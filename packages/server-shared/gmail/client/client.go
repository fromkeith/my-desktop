package client

import (
	"context"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GmailClient(ctx context.Context, accountId string) (*googleClient, error) {
	return GoogleClientFor(ctx, accountId)
}
func GmailClientFor(ctx context.Context, setToBackground bool) (*googleClient, error) {
	accountId := ctx.Value("accountId").(string)
	return GmailClient(ctx, accountId)
}

func (g *googleClient) Bootstrap(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error().
				Ctx(ctx).
				Stack().
				Err(err.(error)).
				Msg("panic bootstraping!")
		}
	}()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Ctx(ctx).
					Stack().
					Err(err.(error)).
					Msg("panic bootstraping people!")
			}
		}()
		err := g.BootstrapPeople(ctx)
		if err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("error bootstrapping people")
		}
	}()
	return g.BootstrapEmail(ctx)
}
