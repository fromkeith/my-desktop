package main

import (
	"context"
	"errors"
	"fmt"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/services/helpers"
	"io"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/genai"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type messageBody struct {
	entry         data.GmailEntry
	body          string
	result        *expectedAnalyzeResult
	src           kafka.Message
	embedding     *genai.ContentEmbedding
	embeddingText string
}
type expectedAnalyzeResult struct {
	Theme      string
	Summary    string
	Categories []string
	Tags       []string
}

var outputDimens int32 = 3072 // its the default, but lets be explict

func main() {
	log.Info().
		Msg("Starting up gemini")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	r := globals.KafkaConsumerGroup("email_injest_available", "gemini")
	defer r.Close()
	dead := globals.KafkaWriter("gemini_dlq")
	defer dead.Close()
	available := globals.KafkaWriter("email_embedding_available")
	defer available.Close()

	ctx := context.WithValue(context.Background(), "service", "gemini")

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
		bodies := make([]messageBody, 0, len(msgs))
		for _, msg := range msgs {
			log.Info().
				Ctx(ctx).
				Str("taskId", string(msg.Key)).
				Msg("processing message")
			var payload data.EmailInjestedPayload
			if err := json.Unmarshal(msg.Value, &payload); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("taskId", string(msg.Key)).
					Msg("failed to unmarshal email")
				failed = append(failed, msg)
				continue
			}
			entry := payload.Entry
			entry.AccountId = payload.AccountId // needed since accountId doesn't marshal to json
			body, err := fetchBody(ctx, entry)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("taskId", string(msg.Key)).
					Msg("failed to fetch message body")
				failed = append(failed, msg)
				continue
			}
			// TODO: strip out replies? but how do we know?
			// TODO: what about attachments?
			asText, err := stripHtml(ctx, *body)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("taskId", string(msg.Key)).
					Msg("failed to strip html")
				failed = append(failed, msg)
				continue
			}
			bodies = append(bodies, messageBody{
				entry: entry,
				body:  asText,
				src:   msg,
			})
		}

		// analyze each body via gemini
		for i, msg := range bodies {
			log.Info().
				Ctx(ctx).
				Str("taskId", msg.entry.ToDocumentId()).
				Int("payloadSize", len(msg.body)).
				Msg("ai-ing document")

			analyzeResult, err := anaylze(ctx, msg)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("taskId", msg.entry.ToDocumentId()).
					Msg("failed to fetch message")
				failed = append(failed, msg.src)
				continue
			}
			bodies[i].result = analyzeResult
		}

		// write the found tags + categories
		if err := writeTagsAndCategories(ctx, bodies); err != nil {
			log.Error().
				Ctx(ctx).
				Stack().
				Err(err).
				Msg("failed to write tags and categories")
		}

		// create the embeddings and assign them to bodies
		toSave := make([]data.EmailSummaryEmbedding, 0, len(bodies))
		if len(bodies) > 0 {
			results, err := createEmbeddings(ctx, bodies)
			if err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("failed to create embeddings")
				for _, msg := range bodies {
					if msg.result != nil {
						failed = append(failed, msg.src)
					}
				}
			} else {
				// assign embeddings to bodies
				var bodyI int = 0
				for _, embd := range results {
					// we expect things to be in the right order
					// but if one failed to analyze, we need to skip it
					// as it won't have an embedding
					for bodies[bodyI].result == nil {
						bodyI++
					}
					msg := bodies[bodyI]
					msg.embedding = embd
					bodies[bodyI] = msg
					// go to next body
					bodyI++

					toSave = append(toSave, data.EmailSummaryEmbedding{
						MessageId: msg.entry.MessageId,
						AccountId: msg.entry.AccountId,
						Embedding: embd.Values,
						Sender:    msg.entry.Sender,
						Receiver:  msg.entry.Receiver,
						Summary:   msg.embeddingText,
						Version:   0,
					})
				}
			}
		}
		// save the embeddings with metadata

		if len(toSave) > 0 {
			data.BulkWriteEmailSummaries(ctx, toSave)
			nextStep := make([]kafka.Message, 0, len(toSave))
			for _, entry := range toSave {
				entryBytes, _ := json.Marshal(entry)
				nextStep = append(nextStep, kafka.Message{
					Key:   []byte(entry.ToDocumentId()),
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

var responseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"Theme": map[string]any{
			"type":        "string",
			"description": "Theme of the email",
		},
		"Summary": map[string]any{
			"type":        "string",
			"description": "1-3 line summary of this email",
		},
		"Categories": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":        "string",
				"description": "A generic category for this email. Must be 1 to 2 words.",
			},
		},
		"Tags": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":        "string",
				"description": "A tag for this email. Must be 1 word.",
			},
		},
	},
	"required": []string{"Theme", "Summary", "Categories", "Tags"},
}

func anaylze(ctx context.Context, email messageBody) (*expectedAnalyzeResult, error) {

	result, err := globals.Gemini().Models.GenerateContent(
		ctx,
		"gemini-flash-latest",
		genai.Text(`Subject: `+email.entry.Subject+`\n\n`+email.body),
		&genai.GenerateContentConfig{
			ResponseMIMEType:   "application/json",
			ResponseJsonSchema: responseSchema,
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{
					Text: `You help categorize & summarize emails. Get out the theme (1 line), a summary (1-3 lines), suggested categories (3-5) and tags (3-8 short tokens, as a string list).
Return as JSON (Theme, Summary, Categories, Tags)`,
				}},
			},
		},
	)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Str("docId", email.entry.ToDocumentId()).
			Msg("failed to generate content")
		return nil, err
	}
	txt := result.Text()
	var res expectedAnalyzeResult
	if err := json.Unmarshal([]byte(txt), &res); err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Str("docId", email.entry.ToDocumentId()).
			Msg("failed to unmarshal analyze result")
		return nil, err
	}
	log.Info().Ctx(ctx).Any("analyzeResult", res).Msg("analyzeResult")
	return &res, nil
}

func createEmbeddings(ctx context.Context, items []messageBody) ([]*genai.ContentEmbedding, error) {
	contents := make([]*genai.Content, 0, len(items))
	for i, item := range items {
		if item.result == nil {
			continue
		}
		res := item.result
		toEmbedd := fmt.Sprintf("%s\n%s\n%s\n%s", res.Theme, res.Summary, strings.Join(res.Categories, ", "), strings.Join(res.Tags, ", "))
		items[i].embeddingText = toEmbedd
		contents = append(contents, genai.NewContentFromText(toEmbedd, genai.RoleUser))
	}

	// https://ai.google.dev/gemini-api/docs/embeddings
	result, err := globals.Gemini().Models.EmbedContent(
		ctx, "gemini-embedding-001",
		contents,
		&genai.EmbedContentConfig{
			TaskType:             "CLUSTERING",
			OutputDimensionality: &outputDimens,
		},
	)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("failed to create embedding")
		return nil, err
	}
	return result.Embeddings, nil
}

func fetchBody(ctx context.Context, entry data.GmailEntry) (*data.GmailEntryBody, error) {
	result := globals.DocDb().Collection("MessageBodies").FindOne(
		ctx,
		bson.M{"_id": entry.ToDocumentId()},
	)
	var body data.GmailEntryBody
	if err := result.Decode(&body); err != nil {
		return nil, err
	}
	return &body, nil
}

func stripHtml(ctx context.Context, body data.GmailEntryBody) (string, error) {
	if body.PlainText != "" {
		return strings.TrimSpace(body.PlainText), nil
	}
	if body.Html != "" {
		// niave limit of body size
		if len(body.Html) > 1024*1024*2 {
			body.Html = body.Html[:1024*1024*2]
		}
		res, err := htmltomarkdown.ConvertString(body.Html)
		return strings.TrimSpace(res), err
	}
	// empty body.. so be e
	return "empty body", nil
}

func writeTagsAndCategories(ctx context.Context, bodies []messageBody) error {
	tagsAndCategories := make([]mongo.WriteModel, 0, len(bodies))
	for _, msg := range bodies {
		if msg.result == nil {
			continue
		}
		// enforce normalization
		for i, tag := range msg.result.Tags {
			msg.result.Tags[i] = strings.TrimSpace(strings.ToLower(tag))
		}
		for i, cat := range msg.result.Categories {
			msg.result.Categories[i] = strings.TrimSpace(strings.ToLower(cat))
		}
		// add to the message itself
		addToSet := bson.M{}
		if len(msg.result.Tags) > 0 {
			addToSet["tags"] = bson.D{{"$each", msg.result.Tags}}
		}
		if len(msg.result.Categories) > 0 {
			addToSet["categories"] = bson.D{{"$each", msg.result.Categories}}
		}
		if len(addToSet) == 0 {
			continue // nothing to add
		}
		tagsAndCategories = append(tagsAndCategories, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{"_id", msg.entry.ToDocumentId()}}).
			SetUpdate(bson.M{
				"$addToSet":    addToSet,
				"$currentDate": bson.M{"updatedAt": true},
				"$inc":         bson.M{"revisionCount": 1},
			}).
			SetUpsert(false),
		)
	}
	col := globals.DocDb().Collection("Messages")
	if _, err := col.BulkWrite(ctx, tagsAndCategories); err != nil {
		return err
	}
	return nil
}
