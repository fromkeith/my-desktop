package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"os"

	"cloud.google.com/go/pubsub/v2"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
)

type gmailPubSubMessagePayload struct {
	EmailAddress string `json:"emailAddress"`
	HistoryId    uint64 `json:"historyId"`
}

func main() {
	log.Info().
		Msg("Starting up Gmail Subscriber")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "service", "gmail-sub"))
	defer cancel()

	go data.StartWriter(ctx)
	go data.StartBodyWriter(ctx)

	client, err := pubsub.NewClient(ctx, os.Getenv("GCLOUD_PROJECT_ID"), option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_SERVICE_ACCOUNT"))))
	if err != nil {
		log.Fatal().Ctx(ctx).Stack().Err(err).Msg("Failed to create client")
	}
	defer client.Close()

	sub := client.Subscriber(
		os.Getenv("GMAIL_PUB_SUB_SUBSCRIPTION"),
	)
	log.Info().Any("sub", sub).Msg("subscriber created")

	err = sub.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		log.Info().Any("m", m).Msg("handling message")
		if err := handleMessage(ctx, m.Data); err != nil {
			log.Error().Ctx(ctx).Stack().Err(err).Msg("Failed to handle message")
		} else {
			m.Ack()
		}
	})
	log.Info().Any("err", err).Msg("recedive returned")
	if err != nil {
		log.Error().Ctx(ctx).Stack().Err(err).Msg("Failed to receive messages")
	}

}

func handleMessage(ctx context.Context, data []byte) error {
	var payload gmailPubSubMessagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	row := globals.Db().QueryRow(
		ctx,
		`SELECT accountId, userId FROM UserEmails WHERE emailAddress = $1`,
		payload.EmailAddress,
	)
	var accountId, userId string
	err := row.Scan(&accountId, &userId)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("no user found for email address %s", payload.EmailAddress)
	} else if err != nil {
		log.Error().Ctx(ctx).Stack().Err(err).Msg("Failed to query for user email")
		return err
	}

	// TODO: let us specify the userId
	client, err := client.GmailClient(ctx, accountId)
	if err != nil {
		return err
	}
	if err := client.SyncEmailUntil(ctx, payload.HistoryId); err != nil {
		return err
	}

	return nil
}
