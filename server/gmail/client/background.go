package client

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var (
	gmailRefreshRequest    = make(chan string, 100)
	contactsRefreshRequest = make(chan string, 100)
)

func CheckForGmailsUpdates(accountId string) {
	gmailRefreshRequest <- accountId
}

func CheckForContactsUpdates(accountId string) {
	contactsRefreshRequest <- accountId
}

type refreshReq struct {
	gmail    bool
	contacts bool
}

func StartBackgroundRefresher(ctx context.Context) {
	// blocks until 100 items read from the queue
	// 5 seconds has passed, or the context is cancelled

	writeWait := make(map[string]refreshReq)

	for {
		select {
		case accountId := <-gmailRefreshRequest:
			if v, ok := writeWait[accountId]; ok {
				v.gmail = true
				writeWait[accountId] = v
			} else {
				writeWait[accountId] = refreshReq{gmail: true}
			}
			if len(writeWait) >= 100 {
				flush(ctx, writeWait)
				clear(writeWait)
			}
		case accountId := <-contactsRefreshRequest:
			if v, ok := writeWait[accountId]; ok {
				v.contacts = true
				writeWait[accountId] = v
			} else {
				writeWait[accountId] = refreshReq{contacts: true}
			}
			if len(writeWait) >= 100 {
				flush(ctx, writeWait)
				clear(writeWait)
			}
		case <-time.After(5 * time.Second):
			flush(ctx, writeWait)
			clear(writeWait)
		case <-ctx.Done():
			return
		}
	}
}

func flush(ctx context.Context, writeWait map[string]refreshReq) {
	if len(writeWait) == 0 {
		return
	}
	nowMinus10 := time.Now().Add(-10 * time.Second)
	for accountId, req := range writeWait {
		var gmailSyncToken string
		if req.gmail {
			row := globals.Db().QueryRow(ctx, `
			SELECT lastSyncTime, historyId
			FROM UserOauthAccounts u
			INNER JOIN GmailSyncStatus g ON g.userId = u.userId
			WHERE u.accountId = $1
			LIMIT 1
				`, accountId)
			var lastUpdate string
			err := row.Scan(&lastUpdate, &gmailSyncToken)
			if err != nil {
				if err != pgx.ErrNoRows {
					log.Error().
						Ctx(ctx).
						Err(err).
						Msg("Failed to scan gmail sync status")
					continue
				}

			}
			if lastUpdate != "" {
				t, err := time.Parse(time.RFC3339, lastUpdate)
				if err == nil {
					// too soon, don't update
					if t.After(nowMinus10) {
						req.gmail = false
					}
				}
			}
		}
		var contactsSyncToken string
		if req.contacts {
			row := globals.Db().QueryRow(ctx, `
			SELECT lastSyncTime, nextSyncToken
			FROM UserOauthAccounts u
			INNER JOIN PeopleSyncStatus g ON g.userId = u.userId
			WHERE u.accountId = $1
			LIMIT 1
				`, accountId)
			var lastUpdate string
			err := row.Scan(&lastUpdate, &contactsSyncToken)
			if err != nil {
				if err != pgx.ErrNoRows {
					log.Error().
						Ctx(ctx).
						Err(err).
						Msg("error scanning people sync status")
				}
				continue
			}
			if lastUpdate != "" {
				t, err := time.Parse(time.RFC3339, lastUpdate)
				if err == nil {
					// too soon, don't update
					if t.After(nowMinus10) {
						req.contacts = false
					}
				}
			}
		}
		// too soon
		if !req.contacts && !req.gmail {
			continue
		}
		client, err := GmailClient(ctx, accountId)
		if err != nil {
			continue
		}

		log.Info().
			Ctx(ctx).
			Bool("gmail", req.gmail).
			Bool("contacts", req.contacts).
			Msg("syncing account")

		if req.gmail {
			client.SyncEmail(ctx, gmailSyncToken)
		}
		if req.contacts {
			client.SyncPeople(ctx, contactsSyncToken)
		}
	}
}
