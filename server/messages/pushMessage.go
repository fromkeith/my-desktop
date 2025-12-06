package messages

import (
	"fromkeith/my-desktop-server/gmail/client"
	"fromkeith/my-desktop-server/gmail/data"
	"fromkeith/my-desktop-server/utils"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/gmail/v1"
)

type PushMessageRow struct {
	NewDocumentState   data.GmailEntry
	AssumedMasterState *data.GmailEntry `json:",omitempty"`
} // @name PushMessageRow

type PushMessageRequest struct {
	Rows []PushMessageRow `validate:"required" json:"rows"`
} // @name PushMessageRequest

type PushMessageResponse struct {
	Conflicts []data.GmailEntry `validate:"required" json:"conflicts"`
} // @name PushMessageResponse

// PushMessage godoc
// @Summary      Update Messages
// @Description  Sync endpoint to push client changes to messages for this account.
// @Tags         email
// @Accept 		 json
// @Param        request body PushMessageRequest true "Push Message Request"
// @Produce      json
// @Success      200  {object}  PushMessageResponse
// @Router       /messages/push [post]
func PushMessage(r *gin.Context) {
	var req PushMessageRequest
	if err := r.ShouldBindBodyWithJSON(&req); err != nil {
		r.JSON(400, gin.H{"error": err.Error()})
		return
	}

	client, err := client.GmailClientFor(r, false)
	if err != nil {
		r.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// only allow updates.. don't allow new documents
	// so we should require that the assumed master state is not nil
	conflicts := make([]data.GmailEntry, 0, 100)
	for _, row := range req.Rows {
		if row.AssumedMasterState == nil {
			r.JSON(400, gin.H{"error": "Missing Assumed Master State"})
			return
		}
		labelNew, labelRemoved := utils.SetDiff(row.AssumedMasterState.Labels, row.NewDocumentState.Labels)
		if len(labelNew) == 0 && len(labelRemoved) == 0 {
			continue
		}
		mod := &gmail.ModifyMessageRequest{
			AddLabelIds:    labelNew,
			RemoveLabelIds: labelRemoved,
		}
		if err := client.UpdateMessage(r, row.AssumedMasterState.MessageId, mod); err != nil {
			r.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	// return conflicts
	r.JSON(200, PushMessageResponse{Conflicts: conflicts})
}
