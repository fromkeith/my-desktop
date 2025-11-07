package people

import (
	"fmt"
	"fromkeith/my-desktop-server/gmail/client"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SyncPeople(r *gin.Context) {
	client, err := client.GmailClientFor(r, true)
	if err != nil {
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get Gmail client: %v", err)})
		return
	}
	err = client.BootstrapPeople(r)
	if err != nil {
		log.Println("failed to bootstrap people:", err)
		r.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to bootstrap people: %v", err)})
		return
	}
	r.JSON(http.StatusOK, gin.H{})

}
