package messages

import (
	"fromkeith/my-desktop-server/globals"
	"fromkeith/my-desktop-server/gmail/data"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func ReInjest(r *gin.Context) {
	messageId := r.Param("messageId")
	userId := r.Param("userId")
	accountId := r.GetString("accountId")

	emailInjest := globals.KafkaWriter("email_injest")
	defer emailInjest.Close()

	value, _ := json.Marshal(data.EmailInjestPayload{
		MessageId: messageId,
		AccountId: accountId,
		UserId:    userId,
	})
	msg := kafka.Message{
		Key:   []byte(accountId + ";" + messageId),
		Value: value,
	}

	if err := emailInjest.WriteMessages(r, msg); err != nil {
		log.Error().
			Ctx(r).
			Err(err).
			Msg("Failed write MessagesAdded to Kafka")
	}
	r.JSON(http.StatusOK, gin.H{})
}
