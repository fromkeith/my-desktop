package main

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/threads"
	"fromkeith/my-desktop-server/utils"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	log.Info().
		Msg("Starting up messageToThread")
	globals.SetupJsonEncoding()
	defer globals.CloseAll()

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), "service", "messageToThread"))
	defer cancel()

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

	streamRes, errChan := utils.BatchMongoStreamChannel(ctx, stream, 10, time.Second)

loop:
	for {
		select {
		case items := <-streamRes:
			handleItems(ctx, items)
		case err := <-errChan:
			log.Error().Ctx(ctx).Stack().Err(err).Msg("error from stream")
			break loop
		case <-ctx.Done():
			break loop
		}
	}

	log.Info().Msg("Exiting")

}
func handleItems(ctx context.Context, items []bson.M) {

	batch := make([]mongo.WriteModel, 0, len(items))

	for _, ev := range items {
		var id string
		if key, ok := ev["documentKey"].(bson.M); ok {
			id = key["_id"].(string)
		} else {
			id = ""
		}
		var operationType string = ev["operationType"].(string)
		switch operationType {
		case "insert", "replace", "update":
			var entry data.GmailEntry
			raw, _ := bson.Marshal(ev["fullDocument"])
			if err := bson.Unmarshal(raw, &entry); err != nil {
				log.Error().
					Ctx(ctx).
					Err(err).
					Any("id", id).
					Msg("failed to unmarshal email in stream")
				continue
			}
			msg := threads.MessageBasic{
				MessageId:    entry.MessageId,
				InternalDate: entry.InternalDate,
				Sender:       entry.Sender,
				Subject:      entry.Subject,
				Snippet:      entry.Snippet,
				Labels:       entry.Labels,
			}
			// create a pipeline to update the thread "messages" item
			// so that we don't have duplicate entries
			pipe := mongo.Pipeline{{
				{
					Key: "$set",
					Value: bson.M{
						"accountId": entry.AccountId,
						"threadId":  entry.ThreadId,
						"messages": bson.M{
							"$let": bson.M{
								"vars": bson.M{
									"newMsg": msg,
								},
								"in": bson.M{
									"$cond": bson.A{
										// CASE 1: messageId already exists in messages.messageId
										bson.M{
											"$in": bson.A{
												"$$newMsg.messageId",
												bson.M{
													"$ifNull": bson.A{
														"$messages.messageId",
														bson.A{}, // default empty array if messages is missing
													},
												},
											},
										},
										// -> replace existing element with newMsg
										bson.M{
											"$map": bson.M{
												"input": bson.M{
													"$ifNull": bson.A{
														"$messages",
														bson.A{},
													},
												},
												"as": "m",
												"in": bson.M{
													"$cond": bson.A{
														bson.M{
															"$eq": bson.A{
																"$$m.messageId",
																"$$newMsg.messageId",
															},
														},
														"$$newMsg", // replace
														"$$m",      // keep old
													},
												},
											},
										},
										// CASE 2: messageId does not exist yet -> append newMsg
										bson.M{
											"$concatArrays": bson.A{
												bson.M{
													"$ifNull": bson.A{
														"$messages",
														bson.A{},
													},
												},
												bson.A{"$$newMsg"},
											},
										},
									},
								},
							},
						},
						// max(mostRecentInternalDate, entry.InternalDate)
						"mostRecentInternalDate": bson.M{
							"$cond": bson.A{
								bson.M{
									"$gt": bson.A{
										entry.InternalDate,
										bson.M{
											"$ifNull": bson.A{
												"$mostRecentInternalDate",
												entry.InternalDate,
											},
										},
									},
								},
								entry.InternalDate,
								bson.M{
									"$ifNull": bson.A{
										"$mostRecentInternalDate",
										entry.InternalDate,
									},
								},
							},
						},
						// thread-level tags/categories as sets
						"tags": bson.M{
							"$setUnion": bson.A{
								bson.M{"$ifNull": bson.A{"$tags", bson.A{}}},
								entry.Tags,
							},
						},
						"categories": bson.M{
							"$setUnion": bson.A{
								bson.M{"$ifNull": bson.A{"$categories", bson.A{}}},
								entry.Categories,
							},
						},
						"updatedAt": "$$NOW",
						"createdAt": bson.M{
							// on insert: createdAt will be $$NOW
							// on update: keep existing createdAt
							"$ifNull": bson.A{"$createdAt", "$$NOW"},
						},
					},
				},
			}}
			batch = append(batch, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": entry.AccountId + ";" + entry.ThreadId}).
				SetUpdate(pipe).
				SetUpsert(true))

		case "delete":
			// ignore for now
		}
	}

	threadCol := globals.DocDb().Collection("MessageThreads")
	_, err := threadCol.BulkWrite(ctx, batch)

	if err != nil {
		log.Error().Ctx(ctx).
			Stack().Err(err).
			Msg("Failed to update thread on message change")
	}
}
