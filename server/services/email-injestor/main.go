package main

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/services/helpers"
	"time"

	"github.com/hibiken/asynq"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/wagslane/go-rabbitmq"
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
			Queues: map[string]int {
				"emails:injest": 10,
			},
			// TODO: create a logger that maps to zerolog
			// Logger: log.,
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("emails:injest", HandleEmailInjest)

	available := globals.KafkaWriter("email_injest_available")
	defer available.Close()

	if err := srv.Run(mux); err != nil {
		log.Fatal().Stack().Err(err).Msg("could not run asynq server")
	}
}

func HandleEmailInjest(ctx context.Context, t *asynq.Task) error {

		entries := make([]data.GmailEntry, 0, len(msgs))
		bodies := make([]data.GmailEntryBody, 0, len(msgs))
		success := make([]rabbitmq.Delivery, 0, len(msgs))
		for _, msg := range msgs {
			log.Info().
				Ctx(ctx).
				Str("taskId", string(msg.MessageId)).
				Msg("processing message")

			entry, body, err := fetchEmail(ctx, msg)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("taskId", string(msg.MessageId)).
					Msg("failed to fetch message")
				if err := helpers.RequeueMessage(ctx, queueName, msg); err != nil {
					log.Error().
						Ctx(ctx).
						Err(err).
						Str("taskId", string(msg.MessageId)).
						Msg("failed to requeue message")
				}
				continue
			}
			entries = append(entries, *entry)
			bodies = append(bodies, *body)
			success = append(success, msg)
		}

		if len(entries) > 0 {
			data.BulkWriteEmailBodies(ctx, bodies)
			data.BulkWriteEmails(ctx, entries)
			nextStep := make([]kafka.Message, 0, len(entries))
			for _, entry := range entries {
				entryBytes, _ := json.Marshal(data.EmailInjestedPayload{
					MessageId: entry.MessageId,
					AccountId: entry.AccountId,
					Entry:     entry,
				})
				nextStep = append(nextStep, kafka.Message{
					Key:   []byte(entry.AccountId + ";" + entry.MessageId),
					Value: entryBytes,
				})
			}
			// make it available to downstream services
			if err := available.WriteMessages(ctx, nextStep...); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("failed to write messages to available topic. Messages lost!")
			}
		}

		for _, s := range success {
			if err := s.Ack(false); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("failed to ack message as success")
			}
		}
	}

}

func fetchEmail(ctx context.Context, msg rabbitmq.Delivery) (*data.GmailEntry, *data.GmailEntryBody, error) {
	// log.Debug().Str("payload", string(msg.Value)).Msg("kafka payload")
	var payload data.EmailInjestPayload
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		return nil, nil, err
	}
	client, err := client.GmailClient(ctx, payload.AccountId)
	if err != nil {
		return nil, nil, err
	}
	return client.FetchGmailEntry(ctx, payload.MessageId)
}
