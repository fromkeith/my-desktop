package main

import (
	"context"
	"errors"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/services/helpers"
	"io"
	"log"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/segmentio/kafka-go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	log.Println("Starting up email-injestor")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	r := globals.KafkaConsumerGroup("email_injest", "fetch")
	defer r.Close()
	dead := globals.KafkaWriter("email_inject_dlq")
	defer dead.Close()
	available := globals.KafkaWriter("email_injest_available")
	defer available.Close()

	ctx := context.Background()

	for {
		log.Println("email-injestor: Waiting for messages")
		msgs, err := helpers.FetchBatch(ctx, r, 10, time.Second)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				log.Println("context canceled; exiting")
				break
			}
			var kerr *kafka.Error
			if errors.As(err, &kerr) && kerr.Temporary() {
				log.Printf("temporary kafka error: %v", err)
				continue
			}
			if errors.Is(err, io.ErrClosedPipe) {
				log.Println("reader closed; exiting")
				break
			}
			log.Printf("fetch error: %v", err)
			continue
		}
		log.Println("email-injestor: Got messages!", len(msgs))

		failed := make([]kafka.Message, 0)
		entries := make([]data.GmailEntry, 0, len(msgs))
		bodies := make([]data.GmailEntryBody, 0, len(msgs))
		for _, msg := range msgs {
			log.Printf("received message (email_injest, fetch): %s", string(msg.Key))
			entry, body, err := fetchEmail(ctx, msg)
			if err != nil {
				log.Println("failed to fetch message:", string(msg.Key), err)
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
					Key:   []byte(entry.MessageId),
					Value: entryBytes,
				})
			}
			// make it available to downstream services
			if err := available.WriteMessages(ctx, nextStep...); err != nil {
				log.Printf("failed to write messages to available topic: %v", err)
			}
		}
		// TODO: order of writing to DLQ vs committing
		// if the DLQ fails? should we just not commit anything?
		if err := r.CommitMessages(ctx, msgs...); err != nil {
			log.Printf("commit failed: %v", err)
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
				log.Printf("failed to write failed messages. Messages may be lost!: %v", err)
			}
		}
	}
}

func fetchEmail(ctx context.Context, msg kafka.Message) (*data.GmailEntry, *data.GmailEntryBody, error) {
	var payload data.EmailInjestPayload
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return nil, nil, err
	}
	client, err := client.GmailClient(ctx, payload.AccountId, false)
	if err != nil {
		return nil, nil, err
	}
	return client.FetchGmailEntry(ctx, payload.MessageId)
}
