package client

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	peopleFields = "addresses,ageRanges,biographies,birthdays,calendarUrls,clientData,coverPhotos,emailAddresses,events,externalIds,genders,imClients,interests,locales,locations,memberships,metadata,miscKeywords,names,nicknames,occupations,organizations,phoneNumbers,photos,relations,sipAddresses,skills,urls,userDefined"
)

func (g *googleClient) SyncPeople(ctx context.Context, syncToken string) error {
	return g.loadPeople(ctx, syncToken)
}

func (g *googleClient) BootstrapPeople(ctx context.Context) error {
	return g.loadPeople(ctx, "")

}
func (g *googleClient) loadPeople(ctx context.Context, syncToken string) error {

	var nextPageToken string = ""

	for {

		req := g.people.People.Connections.List("people/me").
			PersonFields(peopleFields).
			PageSize(1000). // TODO: people will have more than 1000 connections
			RequestSyncToken(true)
		if nextPageToken != "" {
			req.PageToken(nextPageToken)
		}
		if syncToken != "" {
			req.SyncToken(syncToken)
		}
		res, err := req.Do()
		if err != nil {
			return err
		}

		batchWriteModels := make([]mongo.WriteModel, 0, len(res.Connections))
		for _, person := range res.Connections {
			if person.ResourceName == "" {
				continue
			}
			p := data.GooglePerson{
				Person:    *person,
				PersonId:  person.ResourceName,
				AccountId: g.accountId,
			}

			doc := bson.M{}
			b, _ := bson.Marshal(p)
			_ = bson.Unmarshal(b, &doc)
			delete(doc, "updatedAt")
			delete(doc, "revisionCount")
			delete(doc, "createdAt") // let $setOnInsert handle this
			log.Println("person", doc)
			batchWriteModels = append(batchWriteModels, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": p.ToDocumentId()}).
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

		col := globals.DocDb().Collection("People")
		if _, err := col.BulkWrite(ctx, batchWriteModels); err != nil {
			return err
		}
		if res.NextPageToken != "" {
			continue
		}

		_, err = globals.Db().ExecContext(ctx, `
		INSERT INTO people_sync_status (
			user_id,
			next_sync_token,
			last_sync_time
		) VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			next_sync_token = excluded.next_sync_token,
			last_sync_time = MAX(excluded.last_sync_time, last_sync_time)
			`,
			g.userId,
			res.NextSyncToken,
			time.Now().Format(time.RFC3339),
		)
		if err != nil {
			log.Println("Failed to save sync status", err)
		}
		return nil
	}
}
