package main

import (
	"context"
	"errors"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/services/helpers"
	"io"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	log.Info().
		Msg("Starting up email-injestor")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	r := globals.KafkaConsumerGroup("email_injest", "fetch")
	defer r.Close()
	dead := globals.KafkaWriter("email_inject_dlq")
	defer dead.Close()
	available := globals.KafkaWriter("email_injest_available")
	defer available.Close()

	ctx := context.WithValue(context.Background(), "service", "email-injestor")

	for {
		log.Info().
			Ctx(ctx).
			Msg("Waiting for messages")
		msgs, err := helpers.FetchBatch(ctx, r, 10, time.Second)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				log.Info().
					Ctx(ctx).
					Msg("context canceled; exiting")
				break
			}
			var kerr *kafka.Error
			if errors.As(err, &kerr) && kerr.Temporary() {
				log.Warn().
					Ctx(ctx).
					Err(err).
					Msg("temporary kafka error")
				continue
			}
			if errors.Is(err, io.ErrClosedPipe) {
				log.Info().
					Ctx(ctx).
					Err(err).
					Msg("reader closed; exiting")
				break
			}
			log.Printf("fetch error: %v", err)
			continue
		}
		log.Info().
			Ctx(ctx).
			Int("count", len(msgs)).
			Msg("Got messages")

		failed := make([]kafka.Message, 0)
		entries := make([]data.GmailEntry, 0, len(msgs))
		bodies := make([]data.GmailEntryBody, 0, len(msgs))
		for _, msg := range msgs {
			log.Info().
				Ctx(ctx).
				Str("taskId", string(msg.Key)).
				Msg("processing message")

			entry, body, err := fetchEmail(ctx, msg)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("taskId", string(msg.Key)).
					Msg("failed to fetch message")
				failed = append(failed, msg)
				continue
			}
			entries = append(entries, *entry)
			bodies = append(bodies, *body)
		}

		if len(entries) > 0 {
			data.BulkWriteEmailBodies(ctx, bodies)
			data.BulkWriteEmails(ctx, entries)
			nextStep := make([]kafka.Message, 0, len(entries))
			for _, entry := range entries {
				entryBytes, _ := json.Marshal(entry)
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
		// TODO: order of writing to DLQ vs committing
		// if the DLQ fails? should we just not commit anything?
		if err := r.CommitMessages(ctx, msgs...); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("Failed to commit messages")
		}
		if len(failed) > 0 {
			for i, f := range failed {
				// overwrite topic and other metadata
				failed[i] = kafka.Message{
					Key:     f.Key,
					Value:   f.Value,
					Headers: f.Headers,
				}
			}
			if err := dead.WriteMessages(ctx, failed...); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("failed to write messages to dead topic. Messages lost!")
			}
		}
	}
}

func fetchEmail(ctx context.Context, msg kafka.Message) (*data.GmailEntry, *data.GmailEntryBody, error) {
	// log.Debug().Str("payload", string(msg.Value)).Msg("kafka payload")
	var payload data.EmailInjestPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return nil, nil, err
	}
	client, err := client.GmailClient(ctx, payload.AccountId)
	if err != nil {
		return nil, nil, err
	}
	return client.FetchGmailEntry(ctx, payload.MessageId)
}
