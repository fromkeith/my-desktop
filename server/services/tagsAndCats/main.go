package main

import (
	"context"
	"errors"
	"fmt"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	log.Info().
		Msg("Starting up tagsAndCa")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	ctx := context.WithValue(context.Background(), "service", "tagsAndCats")

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "operationType",
				Value: bson.D{{
					Key: "$in", Value: bson.A{"insert", "update", "replace", "delete"}},
				}},
		}}},
	}

	opts := options.ChangeStream().
		SetFullDocument(options.UpdateLookup)
	stream, err := globals.DocDb().Collection("Messages").Watch(ctx, pipeline, opts)
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("Failed to start change stream")
		return
	}
	defer stream.Close(ctx)
	for stream.Next(ctx) {
		var ev bson.M
		if err := stream.Decode(&ev); err != nil {
			// NOTE: must pass &ev — decoding into a nil value triggers "cannot Decode to nil value"
			log.Printf("[labels] decode error: %v", err)
			continue
		}
		var id string
		if key, ok := ev["documentKey"].(bson.M); ok {
			id = key["_id"].(string)
		} else {
			id = ""
		}
		var operationType string = ev["operationType"].(string)
		switch operationType {
		case "insert", "replace", "update":
			var email data.GmailEntry
			raw, _ := bson.Marshal(ev["fullDocument"])
			if err := bson.Unmarshal(raw, &email); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Any("id", id).
					Msg("failed to unmarshal email in stream")
				continue
			}
			if err := syncTagsAndCats(ctx, email); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("docId", email.ToDocumentId()).
					Msg("failed to sync tags and categories")
			}
		case "delete":
			if id == "" {
				continue // don't have a key we can delete with
			}
			if err := deleteTagsAndCats(ctx, id); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Str("docId", id).
					Msg("failed to sync tags and categories")
			}
		}

	}

	if err := stream.Err(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal().Stack().Err(err).Msg("ChangeStream closed")
	}
	log.Info().Msg("Exiting")

}

func deleteTagsAndCats(ctx context.Context, docId string) error {
	parts := strings.Split(docId, ";")
	if len(parts) != 2 {
		return fmt.Errorf("invalid docId format")
	}
	accountId := parts[0]
	messageId := parts[1]
	tagsCol := globals.DocDb().Collection("MessageTags")
	_, err := tagsCol.DeleteMany(ctx, bson.M{"accountId": accountId, "messageId": messageId})
	if err != nil {
		return err
	}
	catCol := globals.DocDb().Collection("MessageCategories")
	_, err = catCol.DeleteMany(ctx, bson.M{"accountId": accountId, "messageId": messageId})
	if err != nil {
		return err
	}
	return nil

}

func syncTagsAndCats(ctx context.Context, email data.GmailEntry) error {
	log.Info().
		Ctx(ctx).
		Str("docId", email.ToDocumentId()).
		Msg("Syncing tags and categories")

	if err := syncTags(ctx, email); err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed syncTags")
		return err
	}
	if err := syncCats(ctx, email); err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed syncCats")
		return err
	}

	return nil
}

func syncTags(ctx context.Context, email data.GmailEntry) error {
	db := globals.DocDb()
	cur, err := db.Collection("MessageTags").Find(ctx, bson.M{"messageId": email.MessageId, "accountId": email.AccountId})
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed to get previous message tags")
		return err
	}
	defer cur.Close(ctx)
	var oldTags []data.MessageTag
	if err := cur.All(ctx, &oldTags); err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed to get previous message tags (decode)")
		return err
	}
	// fill with all tags, we will remove ones we still have
	// then be left with a map of tags we need to remove
	var toRemove = make(map[string]bool)
	for _, tag := range oldTags {
		toRemove[tag.Tag] = true
	}
	toAdd := make(map[string]bool)
	for _, tag := range email.Tags {
		if _, ok := toRemove[tag]; !ok {
			toAdd[tag] = true
		} else {
			// remove from the toRemove list
			delete(toRemove, tag)
		}
	}
	log.Info().
		Ctx(ctx).
		Str("docId", email.ToDocumentId()).
		Any("addTags", toAdd).
		Any("removeTags", toRemove).
		Msg("Modifying Tags")
	toWrite := make([]mongo.WriteModel, 0, len(toAdd)+len(toRemove))
	for t := range toAdd {
		entry := data.MessageTag{
			MessageId: email.MessageId,
			AccountId: email.AccountId,
			Tag:       t,
			Source:    "system", // TODO: how to determine this
		}
		doc := bson.M{}
		b, _ := bson.Marshal(entry)
		_ = bson.Unmarshal(b, &doc)
		delete(doc, "updatedAt")
		delete(doc, "createdAt") // let $setOnInsert handle this

		filter := bson.M{"_id": entry.ToDocumentId()}
		update := bson.M{
			"$setOnInsert": doc,
			"$currentDate": bson.M{"updatedAt": true},
		}
		toWrite = append(toWrite, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}
	for t := range toRemove {
		entry := data.MessageTag{
			MessageId: email.MessageId,
			AccountId: email.AccountId,
			Tag:       t,
		}

		filter := bson.M{"_id": entry.ToDocumentId()}
		toWrite = append(toWrite, mongo.NewDeleteOneModel().
			SetFilter(filter))
	}
	if len(toWrite) == 0 {
		return nil
	}
	_, err = db.Collection("MessageTags").BulkWrite(ctx, toWrite)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed to write tags")
		return err
	}

	// account tags now
	toWrite = toWrite[:0]
	if len(toAdd) > 0 {
		toWrite := make([]mongo.WriteModel, 0, len(toAdd))
		for t := range toAdd {
			entry := data.AccountTag{
				AccountId: email.AccountId,
				Tag:       t,
				CreatedAt: time.Now().UTC(),
			}
			doc := bson.M{}
			b, _ := bson.Marshal(entry)
			_ = bson.Unmarshal(b, &doc)
			delete(doc, "updatedAt")
			delete(doc, "messageCount") // let $inc handle this

			filter := bson.M{"_id": entry.ToDocumentId()}
			update := bson.M{
				"$setOnInsert": doc,
				"$currentDate": bson.M{"updatedAt": true},
				"$inc":         bson.M{"messageCount": 1},
			}
			toWrite = append(toWrite, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true))
		}
	}
	// decrement count
	if len(toRemove) > 0 {
		for t := range toRemove {
			entry := data.AccountTag{
				AccountId: email.AccountId,
				Tag:       t,
			}
			filter := bson.M{"_id": entry.ToDocumentId()}
			update := bson.M{
				"$currentDate": bson.M{"updatedAt": true},
				"$inc":         bson.M{"messageCount": -1},
			}
			toWrite = append(toWrite, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(false))
		}
	}
	_, err = db.Collection("AccountTags").BulkWrite(ctx, toWrite)
	return err
}

func syncCats(ctx context.Context, email data.GmailEntry) error {
	db := globals.DocDb()
	cur, err := db.Collection("MessageCategories").Find(ctx, bson.M{"messageId": email.MessageId, "accountId": email.AccountId})
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed to get previous message tags")
		return err
	}
	defer cur.Close(ctx)
	var oldTags []data.MessageCategory
	if err := cur.All(ctx, &oldTags); err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed to get previous message tags (decode)")
		return err
	}
	// fill with all tags, we will remove ones we still have
	// then be left with a map of tags we need to remove
	var toRemove = make(map[string]bool)
	for _, tag := range oldTags {
		toRemove[tag.Category] = true
	}
	toAdd := make(map[string]bool)
	for _, tag := range email.Categories {
		if _, ok := toRemove[tag]; !ok {
			toAdd[tag] = true
		} else {
			// remove from the toRemove list
			delete(toRemove, tag)
		}
	}
	log.Info().
		Ctx(ctx).
		Str("docId", email.ToDocumentId()).
		Any("addCats", toAdd).
		Any("removeCats", toRemove).
		Msg("Modifying Cats")
	toWrite := make([]mongo.WriteModel, 0, len(toAdd)+len(toRemove))
	for t := range toAdd {
		entry := data.MessageCategory{
			MessageId: email.MessageId,
			AccountId: email.AccountId,
			Category:  t,
			Source:    "system", // TODO: how to determine this
		}
		doc := bson.M{}
		b, _ := bson.Marshal(entry)
		_ = bson.Unmarshal(b, &doc)
		delete(doc, "updatedAt")
		delete(doc, "createdAt") // let $setOnInsert handle this

		filter := bson.M{"_id": entry.ToDocumentId()}
		update := bson.M{
			"$setOnInsert": doc,
			"$currentDate": bson.M{"updatedAt": true},
		}
		toWrite = append(toWrite, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}
	for t := range toRemove {
		entry := data.MessageCategory{
			MessageId: email.MessageId,
			AccountId: email.AccountId,
			Category:  t,
		}

		filter := bson.M{"_id": entry.ToDocumentId()}
		toWrite = append(toWrite, mongo.NewDeleteOneModel().
			SetFilter(filter))
	}
	if len(toWrite) == 0 {
		return nil
	}
	_, err = db.Collection("MessageCategories").BulkWrite(ctx, toWrite)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("docId", email.ToDocumentId()).
			Stack().
			Err(err).
			Msg("Failed to write categories")
		return err
	}

	// account categories now
	toWrite = toWrite[:0]
	if len(toAdd) > 0 {
		for t := range toAdd {
			entry := data.AccountCategory{
				AccountId: email.AccountId,
				Category:  t,
				CreatedAt: time.Now().UTC(),
			}
			doc := bson.M{}
			b, _ := bson.Marshal(entry)
			_ = bson.Unmarshal(b, &doc)
			delete(doc, "updatedAt")
			delete(doc, "messageCount") // let $inc handle this

			filter := bson.M{"_id": entry.ToDocumentId()}
			update := bson.M{
				"$setOnInsert": doc,
				"$currentDate": bson.M{"updatedAt": true},
				"$inc":         bson.M{"messageCount": 1},
			}
			toWrite = append(toWrite, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true))
		}
	}
	// decrement count
	if len(toRemove) > 0 {
		for t := range toRemove {
			entry := data.AccountCategory{
				AccountId: email.AccountId,
				Category:  t,
			}
			filter := bson.M{"_id": entry.ToDocumentId()}
			update := bson.M{
				"$currentDate": bson.M{"updatedAt": true},
				"$inc":         bson.M{"messageCount": -1},
			}
			toWrite = append(toWrite, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(false))
		}
	}
	_, err = db.Collection("AccountCategories").BulkWrite(ctx, toWrite)
	return err
}
