package messages

import (
	"fmt"
	"fromkeith/my-desktop-server/gmail/client"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func ForceSyncMessages(r *gin.Context) {
	client, err := client.GmailClientFor(r, true)
	if err != nil {
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get Gmail client: %v", err)})
		return
	}
	err = client.BootstrapEmail(r)
	if err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("failed to bootstrap messages")
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to bootstrap messages: %v", err)})
		return
	}
	r.JSON(http.StatusOK, gin.H{})

}
