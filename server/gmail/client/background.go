package client

import (
	"context"
	"fromkeith/my-desktop-server/globals"
	"log"
	"time"
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
			row := globals.Db().QueryRow(`
			SELECT last_sync_time, history_id
			FROM user_oauth_accounts u
			INNER JOIN gmail_sync_status g ON g.user_id = u.user_id
			WHERE u.account_id = ?
			LIMIT 1
				`, accountId)
			var lastUpdate string
			err := row.Scan(&lastUpdate, &gmailSyncToken)
			if err != nil {
				log.Println("error scanning gmail sync status:", err)
				continue
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
			row := globals.Db().QueryRow(`
			SELECT last_sync_time, next_sync_token
			FROM user_oauth_accounts u
			INNER JOIN people_sync_status g ON g.user_id = u.user_id
			WHERE u.account_id = ?
			LIMIT 1
				`, accountId)
			var lastUpdate string
			err := row.Scan(&lastUpdate, &contactsSyncToken)
			if err != nil {
				log.Println("error scanning people sync status:", err)
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
		log.Println("syncing ", accountId, "gmail:", req.gmail, "contacts:", req.contacts)
		// too soon
		if !req.contacts && !req.gmail {
			continue
		}
		client, err := GmailClient(ctx, accountId, false)
		if err != nil {
			continue
		}

		if req.gmail {
			client.SyncEmail(ctx, gmailSyncToken)
		}
		if req.contacts {
			client.SyncPeople(ctx, contactsSyncToken)
		}
	}
}
