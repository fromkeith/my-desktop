package client

import (
	"context"
	"encoding/base64"
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"

	"log"
	"strings"
	"sync"

	"net/mail"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type gmailClient struct {
	svc       *gmail.Service
	userId    string
	accountId string
}

func GmailClient(ctx context.Context, accountId string, setToBackground bool) (*gmailClient, error) {
	svc, userId, err := GmailClientForUser(ctx, accountId, setToBackground)
	if err != nil {
		return nil, err
	}
	return &gmailClient{
		svc:       svc,
		userId:    userId,
		accountId: accountId,
	}, nil
}
func GmailClientFor(ctx context.Context, setToBackground bool) (*gmailClient, error) {
	accountId := ctx.Value("accountId").(string)
	return GmailClient(ctx, accountId, setToBackground)
}

// loads the first 500 messages
func (g *gmailClient) Bootstrap(ctx context.Context) error {
	log.Println("Bootstrapping gmail inbox")

	// get this up front, so we have the right history id at the end
	prof, err := g.svc.Users.GetProfile("me").Do()
	if err != nil {
		log.Println("Failed to get user baseline", err)
		return err
	}
	lastHistoryId := prof.HistoryId

	pageToken := ""
	var remaining int64 = 10

	const workers = 8
	idCh := make(chan string, 500)
	errCh := make(chan error, 500)
	maxInternalDateChan := make(chan int64, workers)
	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go g.fetchContentsWorker(ctx, &wg, idCh, errCh, maxInternalDateChan)
	}

	for remaining > 0 {
		pageSize := min(remaining, 100)

		listCall := g.svc.Users.Messages.
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
			idCh <- m.Id
			remaining--
		}

		if listRes.NextPageToken == "" || remaining == 0 {
			break
		}
		pageToken = listRes.NextPageToken
	}
	log.Println("done loading ids")
	close(idCh)
	// wait until we are done, helper
	doneWait := make(chan bool)
	go func() {
		wg.Wait()
		log.Println("done watiting")
		doneWait <- true
	}()
	// waits until done and gets all errors out
	var maxInternalDate int64 = 0
errorLoop:
	for {
		select {
		case err := <-errCh:
			log.Println("Had error fetching", err)
		case maxD := <-maxInternalDateChan:
			maxInternalDate = max(maxD, maxInternalDate)
		case <-doneWait:
			break errorLoop
		}
	}
	log.Println("done fetching?")

	_, err = globals.Db().ExecContext(ctx, `
	INSERT INTO gmail_sync_status (
		user_id,
		history_id,
		until
	) VALUES (?, ?, ?)
	ON CONFLICT(user_id) DO UPDATE SET
		history_id = excluded.history_id,
		until = MAX(excluded.until, until)
		`,
		g.userId,
		lastHistoryId,
		maxInternalDate,
	)
	if err != nil {
		log.Println("Failed to save sync status", err)
	}

	log.Println("done bootstrapping")
	return nil
}

func (g *gmailClient) FetchOneMessage(ctx context.Context, messageId string) error {
	idCh := make(chan string, 1)
	errCh := make(chan error, 1)
	maxInternalDateChan := make(chan int64, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	go g.fetchContentsWorker(ctx, &wg, idCh, errCh, maxInternalDateChan)
	idCh <- messageId
	close(idCh)

	// wait until we are done, helper
	doneWait := make(chan bool)
	go func() {
		wg.Wait()
		log.Println("done watiting")
		doneWait <- true
	}()
	// waits until done and gets all errors out
	var lastError error
errorLoop:
	for {
		select {
		case err := <-errCh:
			lastError = err
		case <-maxInternalDateChan:
			// do nothing
		case <-doneWait:
			break errorLoop
		}
	}
	return lastError

}

func (g *gmailClient) fetchContentsWorker(ctx context.Context, wg *sync.WaitGroup, idCh chan string, errCh chan error, maxInternalDateChan chan int64) {
	defer func() {
		e := recover()
		if e != nil {
			log.Println("panic")
			log.Println(e)
		}
	}()
	defer wg.Done()
	var maxInternalDate int64 = 0
	for id := range idCh {
		log.Println("loading message", id)
		msg, err := g.svc.Users.Messages.
			Get("me", id).
			Format("full").
			Do()
		if err != nil {
			// ignoring not found
			if gErr, ok := err.(*googleapi.Error); ok && gErr.Code == 404 {
				continue
			}
			log.Println("Failed to load message", err)
			errCh <- err
			continue
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
			AdditionalReceivers: map[string][]data.PersonInfo{
				"bcc": peopleFrom(headers, "bcc"),
				"cc":  peopleFrom(headers, "cc"),
			},
		}
		log.Println("Entry", entry)
		maxInternalDate = max(maxInternalDate, entry.InternalDate)
		data.WriteGmailEntry(entry)
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
		data.WriteGmailEntryBody(body)
	}
	log.Println("done fetching contents")
	maxInternalDateChan <- maxInternalDate

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
		log.Printf("decode body failed: %v and %v", err, err2)
		log.Println("----")
		log.Println(s[0:256])
		log.Println("----")
		return ""
	}
	return string(b)
}
