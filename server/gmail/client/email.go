package client

import (
	"context"
	"encoding/base64"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

func (g *googleClient) SyncEmail(ctx context.Context, syncToken string) error {

	emailInjest := globals.KafkaWriter("email_injest")
	defer emailInjest.Close()
	writeQueue := make([]kafka.Message, 0, 50)

	var startHistoryId uint64
	startHistoryId, _ = strconv.ParseUint(syncToken, 10, 64)
	var nextHistoryId uint64 = startHistoryId

	pageToken := ""
	log.Info().
		Ctx(ctx).
		Uint64("startHistory", startHistoryId).
		Msg("SyncEmail")

	for {

		listCall := g.gmail.Users.History.
			List("me").
			StartHistoryId(startHistoryId).
			MaxResults(500)

		if pageToken != "" {
			listCall = listCall.PageToken(pageToken)
		}
		// get a list of messages ids
		listRes, err := listCall.Do()
		if err != nil {
			return err
		}
		nextHistoryId = listRes.HistoryId
		if len(listRes.History) == 0 {
			break
		}
		for _, history := range listRes.History {
			// new messages
			for _, message := range history.MessagesAdded {
				value, _ := json.Marshal(data.EmailInjestPayload{
					MessageId: message.Message.Id,
					AccountId: g.accountId,
					UserId:    g.userId,
				})
				msg := kafka.Message{
					Key:   []byte(g.accountId + ";" + message.Message.Id),
					Value: value,
				}
				writeQueue = append(writeQueue, msg)
				if len(writeQueue) >= 50 {
					if err := emailInjest.WriteMessages(ctx, writeQueue...); err != nil {
						log.Error().
							Ctx(ctx).
							Err(err).
							Msg("Failed write MessagesAdded to Kafka")
						return err
					}
					writeQueue = writeQueue[:0]
				}
			}
			// deleted messages.. not trashed.. actually deleted
			for _, message := range history.MessagesDeleted {
				data.DeleteGmailEntry(g.accountId, message.Message.Id)
			}
			for _, message := range history.LabelsAdded {
				data.UpdateGmailEntryFields(g.accountId, message.Message.Id, bson.M{
					"$addToSet": bson.M{
						"labels": bson.D{{"$each", message.LabelIds}},
					},
				})
			}
			for _, message := range history.LabelsRemoved {
				data.UpdateGmailEntryFields(g.accountId, message.Message.Id, bson.M{
					"$pull": bson.M{
						"labels": bson.D{{"$in", message.LabelIds}},
					},
				})
			}
		}
		if listRes.NextPageToken == "" {
			break
		}
		pageToken = listRes.NextPageToken

	}
	if len(writeQueue) >= 0 {
		if err := emailInjest.WriteMessages(ctx, writeQueue...); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("Failed write MessagesAdded to Kafka (2)")
			return err
		}
	}

	_, err := globals.Db().Exec(ctx, `
	INSERT INTO GmailSyncStatus (
		userId,
		historyId,
		until,
		lastSyncTime
	) VALUES ($1, $2, $3, $4)
	ON CONFLICT(userId) DO UPDATE SET
		historyId = EXCLUDED.historyId,
		until = LEAST(EXCLUDED.until, GmailSyncStatus.until),
		lastSyncTime = GREATEST(EXCLUDED.lastSyncTime, GmailSyncStatus.lastSyncTime)
		`,
		g.userId,
		nextHistoryId,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("Failed save sync status")
	}

	return nil

}

// loads the first 500 messages
func (g *googleClient) BootstrapEmail(ctx context.Context) error {
	log.Info().
		Ctx(ctx).
		Msg("Bootstrapping gmail inbox")

	// get this up front, so we have the right history id at the end
	prof, err := g.gmail.Users.GetProfile("me").Do()
	if err != nil {
		log.Error().
			Ctx(ctx).
			Err(err).
			Msg("Failed to get user baseline")
		return err
	}
	emailInjest := globals.KafkaWriter("email_injest")
	defer emailInjest.Close()
	writeQueue := make([]kafka.Message, 0, 50)

	lastHistoryId := prof.HistoryId

	pageToken := ""
	var remaining int64 = 500

	for remaining > 0 {
		pageSize := min(remaining, 100)

		listCall := g.gmail.Users.Messages.
			List("me").
			IncludeSpamTrash(false).
			MaxResults(pageSize)

		if pageToken != "" {
			listCall = listCall.PageToken(pageToken)
		}

		// get a list of messages ids
		listRes, err := listCall.Do()
		if err != nil {
			return err
		}
		if len(listRes.Messages) == 0 {
			break
		}
		for _, m := range listRes.Messages {
			value, _ := json.Marshal(data.EmailInjestPayload{
				MessageId: m.Id,
				AccountId: g.accountId,
				UserId:    g.userId,
			})
			msg := kafka.Message{
				Key:   []byte(g.accountId + ";" + m.Id),
				Value: value,
			}
			writeQueue = append(writeQueue, msg)
			if len(writeQueue) >= 50 {
				if err := emailInjest.WriteMessages(ctx, writeQueue...); err != nil {
					log.Error().
						Ctx(ctx).
						Err(err).
						Msg("Failed to write Bootstrap Messages to Kafka")
					return err
				}
				writeQueue = writeQueue[:0]
			}
			remaining--
		}

		if listRes.NextPageToken == "" || remaining == 0 {
			break
		}
		pageToken = listRes.NextPageToken
	}
	if len(writeQueue) > 0 {
		if err := emailInjest.WriteMessages(ctx, writeQueue...); err != nil {
			log.Error().
				Ctx(ctx).
				Err(err).
				Msg("Failed to write Bootstrap Messages to Kafka (2)")
			return err
		}
	}

	_, err = globals.Db().Exec(ctx, `
	INSERT INTO GmailSyncStatus (
		userId,
		historyId,
		until,
		lastSyncTime
	) VALUES ($1, $2, $3, $4)
	ON CONFLICT(userId) DO UPDATE SET
		historyId = EXCLUDED.historyId,
		until = LEAST(EXCLUDED.until, GmailSyncStatus.until),
		lastSyncTime = GREATEST(EXCLUDED.lastSyncTime, GmailSyncStatus.lastSyncTime)
		`,
		g.userId,
		lastHistoryId,
		time.Now().UTC(),
		time.Now().UTC(),
	)
	if err != nil {
		log.Error().
			Ctx(ctx).
			Str("accountId", g.accountId).
			Str("userId", g.userId).
			Err(err).
			Msg("Failed to save sync status in bootstrap")
	}

	return nil
}

func (g *googleClient) FetchGmailEntry(ctx context.Context, id string) (*data.GmailEntry, *data.GmailEntryBody, error) {

	log.Info().
		Ctx(ctx).
		Str("messageId", id).
		Msg("FetchGmailEntry")

	msg, err := g.gmail.Users.Messages.
		Get("me", id).
		Format("full").
		Do()
	if err != nil {
		// ignoring not found
		if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
			return nil, nil, gErr
		}
		log.Error().
			Ctx(ctx).
			Str("messageId", id).
			Err(err).
			Msg("Failed to load message")
		return nil, nil, err
	}
	headers := headerMap(msg.Payload.Headers)
	var replyTo *data.PersonInfo
	r := personFrom(headers, "reply-to")
	if r.Email != "" {
		replyTo = &r
	}

	entry := data.GmailEntry{
		UserId:       g.userId,
		AccountId:    g.accountId,
		MessageId:    msg.Id,
		ThreadId:     msg.ThreadId,
		Labels:       msg.LabelIds,
		Subject:      headers["subject"],
		Snippet:      msg.Snippet,
		HistoryId:    msg.HistoryId,
		InternalDate: msg.InternalDate,
		Headers:      headers,
		Sender:       personFrom(headers, "from"),
		Receiver:     peopleFrom(headers, "to"),
		ReceivedAt:   headers["date"],
		ReplyTo:      replyTo,
		IsDeleted:    false,
		AdditionalReceivers: map[string][]data.PersonInfo{
			"bcc": peopleFrom(headers, "bcc"),
			"cc":  peopleFrom(headers, "cc"),
		},
		// set as empty so they atleast exist
		Tags:       make([]string, 0),
		Categories: make([]string, 0),
	}

	text, html, hasAtt, inlineIds := extractBodies(msg.Payload)
	hasAttInt := 0
	if hasAtt {
		hasAttInt = 1
	}
	body := data.GmailEntryBody{
		UserId:         entry.UserId,
		MessageId:      entry.MessageId,
		AccountId:      entry.AccountId,
		PlainText:      text,
		Html:           html,
		HasAttachments: hasAttInt,
		AttachmentIds:  inlineIds,
	}
	return &entry, &body, nil

}

func peopleFrom(headers map[string]string, field string) []data.PersonInfo {
	entry, ok := headers[field]
	if !ok {
		return make([]data.PersonInfo, 0)
	}
	addrs, err := mail.ParseAddressList(entry)
	if err != nil {
		return make([]data.PersonInfo, 0)
	}
	out := make([]data.PersonInfo, 0, len(addrs))
	for _, addr := range addrs {
		out = append(out, data.PersonInfo{
			Email: addr.Address,
			Name:  addr.Name,
		})
	}
	return out
}

func personFrom(headers map[string]string, field string) data.PersonInfo {
	entry, ok := headers[field]
	if !ok {
		return data.PersonInfo{}
	}

	addr, err := mail.ParseAddress(entry)
	if err != nil {
		return data.PersonInfo{}
	}
	return data.PersonInfo{
		Email: addr.Address,
		Name:  addr.Name,
	}
}

func headerMap(hs []*gmail.MessagePartHeader) map[string]string {
	m := make(map[string]string, len(hs))
	for _, h := range hs {
		k := strings.ToLower(h.Name)
		if cur, ok := m[k]; ok && cur != "" {
			m[k] = cur + ", " + h.Value
		} else {
			m[k] = h.Value
		}
	}
	return m
}

// extractBodies walks the MIME tree to find best-effort text/plain and text/html.
// Prefers multipart/alternative selection when present.
// Returns decoded UTF-8 strings. (Base64URL decoded; no charset transcoding here.)
func extractBodies(p *gmail.MessagePart) (text string, html string, hasAttachments bool, inlineIDs []string) {
	if p == nil {
		return
	}
	// log.Println("::: MimeType", p.MimeType)
	// for h := range p.Headers {
	// 	log.Println("headers ", h, p.Headers[h].Name, p.Headers[h].Value)
	// }

	switch {
	case strings.HasPrefix(p.MimeType, "multipart/"):
		// If multipart/alternative, prefer the "best" version:
		if strings.EqualFold(p.MimeType, "multipart/alternative") {
			// First collect candidates
			var tCandidate, hCandidate string
			for _, part := range p.Parts {
				t, h, att, inlines := extractBodies(part)
				hasAttachments = hasAttachments || att
				if len(inlines) > 0 {
					inlineIDs = append(inlineIDs, inlines...)
				}
				if t != "" && tCandidate == "" {
					tCandidate = t
				}
				if h != "" {
					hCandidate = h // prefer last html if multiple
				}
			}
			// multipart/alternative prefers HTML if present; else text
			if hCandidate != "" {
				html = hCandidate
			} else {
				text = tCandidate
			}
			return
		}

		// Generic multipart: union of child results; keep first text, first/last html
		for _, part := range p.Parts {
			t, h, att, inlines := extractBodies(part)
			hasAttachments = hasAttachments || att
			if len(inlines) > 0 {
				inlineIDs = append(inlineIDs, inlines...)
			}
			if text == "" && t != "" {
				text = t
			}
			if h != "" {
				html = h // last html wins
			}
		}
		return

	default:
		mt := strings.ToLower(p.MimeType)
		switch mt {
		case "text/plain":
			text = decodeB64URL(p.Body.Data)
		case "text/html":
			html = decodeB64URL(p.Body.Data)
		default:
			// Mark attachments/inline (non-text) parts
			if p.Body != nil && p.Body.AttachmentId != "" {
				hasAttachments = true
				// inline images often have Content-Id header
				for _, h := range p.Headers {
					if strings.EqualFold(h.Name, "Content-Id") {
						inlineIDs = append(inlineIDs, p.Body.AttachmentId)
						break
					}
				}
			}
		}
		return
	}
}

func decodeB64URL(s string) string {
	if s == "" {
		return ""
	}
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		// Some payloads contain standard base64; fall back
		b2, err2 := base64.StdEncoding.DecodeString(s)
		if err2 == nil {
			return string(b2)
		}
		log.Warn().
			Err(err).
			Err(err2).
			Str("bodySnippet", s[0:256]).
			Msg("decode body failed")
		return ""
	}
	return string(b)
}

func (g *googleClient) UpdateMessage(ctx context.Context, messageId string, modifyReq *gmail.ModifyMessageRequest) error {
	_, err := g.gmail.Users.Messages.Modify(g.userId, messageId, modifyReq).Context(ctx).Do()
	return err
}

func (g *googleClient) BulkUpdateMessages(ctx context.Context, batchReq *gmail.BatchModifyMessagesRequest) error {
	err := g.gmail.Users.Messages.BatchModify(g.userId, batchReq).Context(ctx).Do()
	return err
}
