package data

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TODO: This should use a queue system instead of channels and a worker.
// Its job is to bulkWrite to mongoDB.
// This also has no error retrying or proper handling

var (
	writerQueue = make(chan GmailEntry, 512)
	modifyQueue = make(chan modifyGmailEntry, 512)
	bodyQueue   = make(chan GmailEntryBody, 512)
)

type modifyGmailEntry struct {
	AccountId string
	MessageId string
	Data      *bson.M
	Action    string
}

func WriteGmailEntry(entry GmailEntry) {
	writerQueue <- entry
}

func DeleteGmailEntry(accountId, messageId string) {
	modifyQueue <- modifyGmailEntry{
		AccountId: accountId,
		MessageId: messageId,
		Action:    "delete",
	}
}
func UpdateGmailEntryFields(accountId, messageId string, fields bson.M) {
	modifyQueue <- modifyGmailEntry{
		AccountId: accountId,
		MessageId: messageId,
		Action:    "update",
		Data:      &fields,
	}
}

func WriteGmailEntryBody(entry GmailEntryBody) {
	bodyQueue <- entry
}

func StartWriter(ctx context.Context) {
	// blocks until 100 items read from the queue
	// 5 seconds has passed, or the context is cancelled

	writeWait := make([]GmailEntry, 0, 100)
	modifyWait := make([]modifyGmailEntry, 0, 100)

	for {
		select {
		case entry := <-writerQueue:
			writeWait = append(writeWait, entry)
			if len(writeWait) == 100 {
				if err := BulkWriteEmails(ctx, writeWait); err != nil {
					log.Error().
						Ctx(ctx).
						Err(err).
						Msg("error writing entries")
				}
				writeWait = writeWait[:0]
			}
		case modifyEntry := <-modifyQueue:
			modifyWait = append(modifyWait, modifyEntry)
			if len(modifyWait) == 100 {
				if err := bulkModifyEmails(ctx, modifyWait); err != nil {
					log.Error().
						Ctx(ctx).
						Err(err).
						Msg("error modifying entries")
				}
				modifyWait = modifyWait[:0]
			}
		case <-time.After(5 * time.Second):
			if err := BulkWriteEmails(ctx, writeWait); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("error writing entries")
			}
			writeWait = writeWait[:0]
			if err := bulkModifyEmails(ctx, modifyWait); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("error modifying entries")
			}
			modifyWait = modifyWait[:0]
		case <-ctx.Done():
			return
		}
	}
}

func bulkModifyEmails(ctx context.Context, entries []modifyGmailEntry) error {
	if len(entries) == 0 {
		return nil
	}
	batchWriteModels := make([]mongo.WriteModel, 0, len(entries))
	for _, task := range entries {
		if task.Action == "delete" {
			entry := GmailEntry{
				MessageId: task.MessageId,
				AccountId: task.AccountId,
				IsDeleted: true,
				// leaving everything blank to mark as deleted
			}
			doc := bson.M{}
			b, _ := bson.Marshal(entry)
			_ = bson.Unmarshal(b, &doc)
			delete(doc, "updatedAt")
			delete(doc, "revisionCount")
			delete(doc, "createdAt") // do not want to set this
			batchWriteModels = append(batchWriteModels, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": entry.ToDocumentId()}).
				SetUpdate(bson.M{
					"$set":         doc,
					"$currentDate": bson.M{"updatedAt": true},
					"$inc":         bson.M{"revisionCount": 1},
				}).
				SetUpsert(false),
			)
		} else if task.Action == "update" {
			doc := *task.Data
			if v, ok := doc["$set"].(bson.M); ok {
				delete(v, "updatedAt")
				delete(v, "revisionCount")
				delete(v, "createdAt") // do not want to set this
				doc["$set"] = v
			}
			if v, ok := doc["$currentDate"].(bson.M); ok {
				v["updatedAt"] = true
				doc["$currentDate"] = v
			} else {
				doc["$currentDate"] = bson.M{"updatedAt": true}
			}
			if v, ok := doc["$inc"].(bson.M); ok {
				v["revisionCount"] = 1
				doc["$inc"] = v
			} else {
				doc["$inc"] = bson.M{"revisionCount": 1}
			}

			entry := GmailEntry{
				MessageId: task.MessageId,
				AccountId: task.AccountId,
			}
			batchWriteModels = append(batchWriteModels, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": entry.ToDocumentId()}).
				SetUpdate(doc).
				SetUpsert(false),
			)
		}
	}
	col := globals.DocDb().Collection("Messages")
	if _, err := col.BulkWrite(ctx, batchWriteModels); err != nil {
		return err
	}
	return nil
}

func BulkWriteEmails(ctx context.Context, entries []GmailEntry) error {
	if len(entries) == 0 {
		return nil
	}
	// bulk writes the entries to mongoDB
	// updates/writes over existing entries.
	batchWriteModels := make([]mongo.WriteModel, 0, len(entries))
	for _, entry := range entries {
		// if updating.. needs to increment the version in the database
		// also need to remove fields that we are setting when writing
		doc := bson.M{}
		b, _ := bson.Marshal(entry)
		_ = bson.Unmarshal(b, &doc)
		delete(doc, "updatedAt")
		delete(doc, "revisionCount")
		delete(doc, "createdAt") // let $setOnInsert handle this

		batchWriteModels = append(batchWriteModels, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": entry.ToDocumentId()}).
			SetUpdate(bson.M{
				"$set":         doc,
				"$currentDate": bson.M{"updatedAt": true},
				"$setOnInsert": bson.M{
					"createdAt": time.Now(),
				},
				"$inc": bson.M{"revisionCount": 1},
			}).
			SetUpsert(true),
		)
	}
	col := globals.DocDb().Collection("Messages")
	if _, err := col.BulkWrite(ctx, batchWriteModels); err != nil {
		return err
	}
	return nil
}

func StartBodyWriter(ctx context.Context) {
	// blocks until 100 items read from the queue
	// 5 seconds has passed, or the context is cancelled

	writeWait := make([]GmailEntryBody, 0, 100)

	for {
		select {
		case entry := <-bodyQueue:
			writeWait = append(writeWait, entry)
			if len(writeWait) == 100 {
				if err := BulkWriteEmailBodies(ctx, writeWait); err != nil {
					log.Error().
						Ctx(ctx).
						Err(err).
						Msg("error writing entries (bodies)")
				}
				writeWait = writeWait[:0]
			}
		case <-time.After(5 * time.Second):
			if err := BulkWriteEmailBodies(ctx, writeWait); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Msg("error writing entries (bodies)")
			}
			writeWait = writeWait[:0]
		case <-ctx.Done():
			return
		}
	}
}

func BulkWriteEmailBodies(ctx context.Context, entries []GmailEntryBody) error {
	if len(entries) == 0 {
		return nil
	}
	// bulk writes the entries to mongoDB
	// updates/writes over existing entries.
	batchWriteModels := make([]mongo.WriteModel, 0, len(entries))
	for _, entry := range entries {
		// if updating.. needs to increment the version in the database
		doc := bson.M{}
		b, _ := bson.Marshal(entry)
		_ = bson.Unmarshal(b, &doc)
		delete(doc, "updatedAt")
		delete(doc, "revisionCount")
		delete(doc, "createdAt") // let $setOnInsert handle this
		batchWriteModels = append(batchWriteModels, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": entry.ToDocumentId()}).
			SetUpdate(bson.M{
				"$set":         doc,
				"$currentDate": bson.M{"updatedAt": true},
				"$setOnInsert": bson.M{
					"createdAt": time.Now(),
				},
				"$inc": bson.M{"revisionCount": 1},
			}).
			SetUpsert(true),
		)
	}
	col := globals.DocDb().Collection("MessageBodies")
	if _, err := col.BulkWrite(ctx, batchWriteModels); err != nil {
		return err
	}
	return nil
}

func BulkWriteEmailSummaries(ctx context.Context, entries []EmailSummaryEmbedding) error {
	if len(entries) == 0 {
		return nil
	}
	// bulk writes the entries to mongoDB
	// updates/writes over existing entries.
	batchWriteModels := make([]mongo.WriteModel, 0, len(entries))
	for _, entry := range entries {
		// if updating.. needs to increment the version in the database
		doc := bson.M{}
		b, _ := bson.Marshal(entry)
		_ = bson.Unmarshal(b, &doc)
		delete(doc, "updatedAt")
		delete(doc, "createdAt") // let $setOnInsert handle this
		batchWriteModels = append(batchWriteModels, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": entry.ToDocumentId()}).
			SetUpdate(bson.M{
				"$set":         doc,
				"$currentDate": bson.M{"updatedAt": true},
				"$setOnInsert": bson.M{
					"createdAt": time.Now(),
				},
			}).
			SetUpsert(true),
		)
	}
	col := globals.DocDb().Collection("MessageSummaries")
	if _, err := col.BulkWrite(ctx, batchWriteModels); err != nil {
		return err
	}
	return nil
}
