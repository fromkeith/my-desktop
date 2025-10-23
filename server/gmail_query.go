package main

import (
	"context"
	"database/sql"
	"fromkeith/my-desktop-server/globals"
	gmail_client "fromkeith/my-desktop-server/gmail"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListInbox godoc
// @Summary      List the inbox
// @Description  List the user's email inbox
// @Tags         email
// @Produce      json
// @Success      200  {object}  []gmail_client.GmailEntry
// @Router       /gmail/inbox [get]
func ListInbox(r *gin.Context) {

	rows, err := globals.Db().QueryContext(r, `
		SELECT
			g.user_id,
			g.message_id,
			g.thread_id,
			json(g.labels),
			g.subject,
			g.snippet,
			g.history_id,
			g.internal_date,
			json(g.headers),
			json(g.sender),
			json(g.receiver),
			g.received_at,
			g.reply_to,
			json(g.additional_receivers)
		FROM user_oauth_accounts u
		INNER JOIN gmail_entries g ON g.user_id = u.user_id
		WHERE u.account_id = ?
		ORDER BY g.internal_date DESC
		LIMIT 100
		`,
		r.GetString("accountId"),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			bootstrap(r)
			return
		}
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
		return
	}

	res := make([]gmail_client.GmailEntry, 0, 100)
	for rows.Next() {
		var entry gmail_client.GmailEntry
		var labelsJson, headersJson, senderJson, receiverJson, additionalReceiversJson []byte
		err := rows.Scan(
			&entry.UserId,
			&entry.MessageId,
			&entry.ThreadId,
			&labelsJson,
			&entry.Subject,
			&entry.Snippet,
			&entry.HistoryId,
			&entry.InternalDate,
			&headersJson,
			&senderJson,
			&receiverJson,
			&entry.ReceivedAt,
			&entry.ReplyTo,
			&additionalReceiversJson,
		)
		if err != nil {
			log.Println("failed to unmarshal gmail entry", err)
			continue
		}
		json.Unmarshal((labelsJson), &entry.Labels)
		json.Unmarshal((headersJson), &entry.Headers)
		json.Unmarshal((senderJson), &entry.Sender)
		json.Unmarshal((receiverJson), &entry.Receiver)
		json.Unmarshal((additionalReceiversJson), &entry.AdditionalReceivers)
		res = append(res, entry)
	}
	// TODO: this is not a good place if we decide to add filtering
	// to know if we need to sync them or not
	if len(res) == 0 {
		bootstrap(r)
	}

	r.JSON(http.StatusOK, res)
}

func bootstrap(r *gin.Context) {
	log.Println("Need to bookstrap")
	client, err := gmail_client.GmailClientFor(r, true)
	if err != nil {
		log.Println(err)
		r.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to get gmail client"})
		return
	}
	go client.Boostrap(context.Background())
	r.JSON(http.StatusOK, make([]string, 0))
}
