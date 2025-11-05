package messages

import (
	"fromkeith/my-desktop-server/gmail/data"

	"github.com/gin-gonic/gin"
)

func toDocumentId(entry data.GmailEntry) string {
	return entry.ToDocumentId()
}

func toDocumentIdRequest(r *gin.Context, messageId string) string {
	return toDocumentId(data.GmailEntry{
		AccountId: r.GetString("accountId"),
		MessageId: messageId,
	})
}
