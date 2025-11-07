package people

import (
	"fromkeith/my-desktop-server/gmail/data"

	"github.com/gin-gonic/gin"
)

func toDocumentIdRequest(r *gin.Context, personId string) string {
	return (data.GooglePerson{
		AccountId: r.GetString("accountId"),
		PersonId:  personId,
	}).ToDocumentId()
}
