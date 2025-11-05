package data

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// TODO: This should use a queue system instead of channels and a worker.
// Its job is to bulkWrite to mongoDB.
// This also has no error retrying or proper handling

var (
	writerQueue = make(chan GmailEntry, 512)
	bodyQueue   = make(chan GmailEntryBody, 512)
)

func WriteGmailEntry(entry GmailEntry) {
	writerQueue <- entry
}

func WriteGmailEntryBody(entry GmailEntryBody) {
	bodyQueue <- entry
}

func WaitForBodyWrite(accountId, messageId string, timeout time.Duration) {
	select {
	case <-time.After(timeout):
		return
	case <-bodyQueue:
		log.Printf("body written for account %s and message %s", accountId, messageId)
	}
}

func StartWriter(ctx context.Context) {
	// blocks until 100 items read from the queue
	// 5 seconds has passed, or the context is cancelled

	writeWait := make([]GmailEntry, 0, 100)

	for {
		select {
		case entry := <-writerQueue:
			writeWait = append(writeWait, entry)
			if len(writeWait) == 100 {
				if err := bulkWriteEmails(ctx, writeWait); err != nil {
					log.Printf("error writing entries: %v", err)
				}
				writeWait = writeWait[:0]
			}
		case <-time.After(5 * time.Second):
			if err := bulkWriteEmails(ctx, writeWait); err != nil {
				log.Printf("error writing entries: %v", err)
			}
			writeWait = writeWait[:0]
		case <-ctx.Done():
			return
		}
	}
}

func bulkWriteEmails(ctx context.Context, entries []GmailEntry) error {
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
				if err := bulkWriteEmailBodies(ctx, writeWait); err != nil {
					log.Printf("error writing entries: %v", err)
				}
				writeWait = writeWait[:0]
			}
		case <-time.After(5 * time.Second):
			if err := bulkWriteEmailBodies(ctx, writeWait); err != nil {
				log.Printf("error writing entries: %v", err)
			}
			writeWait = writeWait[:0]
		case <-ctx.Done():
			return
		}
	}
}

func bulkWriteEmailBodies(ctx context.Context, entries []GmailEntryBody) error {
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
