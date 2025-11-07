package client

import (
	"context"
	"log"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func GmailClient(ctx context.Context, accountId string, setToBackground bool) (*googleClient, error) {
	return GoogleClientFor(ctx, accountId, setToBackground)
}
func GmailClientFor(ctx context.Context, setToBackground bool) (*googleClient, error) {
	accountId := ctx.Value("accountId").(string)
	return GmailClient(ctx, accountId, setToBackground)
}

func (g *googleClient) Bootstrap(ctx context.Context) error {
	go func() {
		err := g.BootstrapPeople(ctx)
		if err != nil {
			log.Println("error bootstrapping people:", err)
		}
	}()
	return g.BootstrapEmail(ctx)
}
