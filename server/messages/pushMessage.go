package messages

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PushRow struct {
	NewDocumentState   data.GmailEntry
	AssumedMasterState *data.GmailEntry `json:",omitempty"`
}
type PushMessageRequest struct {
	Rows []PushRow `json:"rows"`
}

type PushMessageResponse struct {
	Conflicts []data.GmailEntry `json:"conflicts"`
}

func PushMessage(r *gin.Context) {
	var req PushMessageRequest
	if err := r.ShouldBindBodyWithJSON(&req); err != nil {
		r.JSON(400, gin.H{"error": err.Error()})
		return
	}

	conflicts := make([]data.GmailEntry, 0, 100)

	accountId := r.GetString("accountId")
	batchId := uuid.New().String()
	batchWriteModels := make([]mongo.WriteModel, 0, len(req.Rows))
	expectedUpdates := make(map[string]bool)
	expectedInserts := make(map[string]bool)

	for _, row := range req.Rows {

		filter := bson.M{}
		id := toDocumentId(row.NewDocumentState)
		// for our upsert adjust the filter
		doUpsert := false
		if row.AssumedMasterState != nil {
			// existing doc, with known previous state
			filter["updatedAt"] = row.AssumedMasterState.UpdatedAt
			filter["revisionCount"] = row.AssumedMasterState.RevisionCount
			filter["_id"] = id
			expectedUpdates[id] = true
		} else {
			doUpsert = true
			filter["_id"] = id
			filter["updatedAt"] = bson.M{"$exists": false}
			expectedInserts[id] = true
		}

		// overwrite these fields
		row.NewDocumentState.UpdatedAt = time.Now() // move to server timestamp
		row.NewDocumentState.LastBatchWriteId = batchId
		row.NewDocumentState.AccountId = accountId
		update := bson.M{
			"$set": row.NewDocumentState,
			"$setOnInsert": bson.M{
				"createdAt": time.Now().UnixNano(),
			},
		}
		// set our update
		m := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(doUpsert)
		batchWriteModels = append(batchWriteModels, m)
	}

	db := globals.DocDb()

	coll := db.Collection("Messages")

	client := db.Client()
	session, err := client.StartSession()
	if err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer session.EndSession(r)
	_, err = session.WithTransaction(r, func(ctx context.Context) (interface{}, error) {
		bwRes, err := coll.BulkWrite(ctx, batchWriteModels, options.BulkWrite().SetOrdered(false))
		if err != nil {
			// We do not expect unique index errors because the filters avoid inserting when a row exists.
			// But if a true race inserts between our filter evaluation and write, Mongo may raise a duplicate key.
			// In that case, we treat those as conflicts by ignoring the error and proceeding to compute touched docs.
			// If you prefer, parse err.(mongo.BulkWriteException) and classify duplicates into out.ConflictKeys.
			if bwe, ok := err.(mongo.BulkWriteException); ok {
				_ = bwe // proceed; we will compute conflicts by marker
			} else {
				return nil, err
			}
		}
		// Collect docs that were inserted
		if len(bwRes.UpsertedIDs) > 0 {
			for _, uid := range bwRes.UpsertedIDs {
				id := uid.(string)
				if _, ok := expectedUpdates[id]; ok {
					delete(expectedUpdates, id)
				} else if _, ok := expectedInserts[id]; ok {
					delete(expectedInserts, id)
				}
			}
		}
		// look for our expected updates
		// populating conflicts with the data of the documents that were not updated
		if len(expectedUpdates) > 0 {
			var checkForUpdateIds = make([]string, 0, len(expectedUpdates))
			for id := range expectedUpdates {
				checkForUpdateIds = append(checkForUpdateIds, id)
			}
			// Find the docs we failed to update, because they had a conflict
			cur, err := coll.Find(
				ctx,
				bson.M{"lastBatchWriteId": bson.M{"$ne": batchId}, "_id": bson.M{"$in": checkForUpdateIds}},
			)
			if err != nil {
				return nil, err
			}
			defer cur.Close(ctx)

			err = cur.All(ctx, &conflicts)
			if err != nil {
				return nil, err
			}
		}
		if len(expectedInserts) > 0 {
			// We expected to insert these items... but failed.
			// Probably because they exist already? Let's find them and append
			// them to our conflicts
			var checkForExisting = make([]string, 0, len(expectedInserts))
			for id := range expectedInserts {
				checkForExisting = append(checkForExisting, id)
			}
			filter := bson.M{
				"messageId": bson.M{"$in": checkForExisting},
				"accountId": accountId,
			}
			cur, err := coll.Find(
				ctx,
				filter,
			)
			if err != nil {
				return nil, err
			}
			defer cur.Close(ctx)

			conflicts2 := make([]data.GmailEntry, 0, 100)
			err = cur.All(ctx, &conflicts2)
			if err != nil {
				return nil, err
			}
			for _, conflict := range conflicts2 {
				delete(expectedInserts, toDocumentId(conflict))
				conflicts = append(conflicts, conflict)
			}
			// TODO: if expectedInserts is not empty, what should we do?
		}
		return nil, nil
	})
	if err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}
	r.JSON(http.StatusOK, PushMessageResponse{
		Conflicts: conflicts,
	})
}
