package main

import (
	"context"
	"fromkeith/my-desktop-server/shared/globals"
	"fromkeith/my-desktop-server/shared/gmail/client"
	"fromkeith/my-desktop-server/shared/gmail/data"
	"fromkeith/my-desktop-server/shared/helpers"

	"github.com/hibiken/asynq"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	queueName = "injestor"
)

func main() {
	log.Info().
		Msg("Starting up email-injestor")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	ctx := context.WithValue(context.Background(), "service", "email-injestor")

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: globals.AsyncAddr(),
		},
		asynq.Config{
			Concurrency: 10,
			BaseContext: func() context.Context {
				return ctx
			},
			// which queues we want to listen to
			Queues: map[string]int{
				"emails:injest": 10,
			},
			// TODO: create a logger that maps to zerolog
			// Logger: log.,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("emails:injest", HandleEmailInjest)

	if err := srv.Run(mux); err != nil {
		log.Fatal().Stack().Err(err).Msg("could not run asynq server")
	}
}

func HandleEmailInjest(ctx context.Context, t *asynq.Task) error {
	log.Info().
		Ctx(ctx).
		Str("taskId", t.Type()).
		RawJSON("payload", t.Payload()).
		Msg("processing message")

	entry, body, err := fetchEmail(ctx, t)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Str("taskId", t.Type()).
			Msg("failed to fetch message")
		// Requeue with backoff and return nil so asynq doesn't retry.
		if err := helpers.RequeueTaskWithBackoff(ctx, t); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Str("taskId", t.Type()).
				Msg("failed to requeue message")
		}
		// We've handled the retry manually, return nil to ACK the message to asynq
		return nil
	}

	data.BulkWriteEmailBodies(ctx, []data.GmailEntryBody{*body})
	data.BulkWriteEmails(ctx, []data.GmailEntry{*entry})

	return nil
}

func fetchEmail(ctx context.Context, task *asynq.Task) (*data.GmailEntry, *data.GmailEntryBody, error) {
	var payload data.EmailInjestPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, nil, err
	}
	client, err := client.GmailClient(ctx, payload.AccountId)
	if err != nil {
		return nil, nil, err
	}
	return client.FetchGmailEntry(ctx, payload.MessageId)
}
