package data

import "time"

type EmailSummaryEmbedding struct {
	AccountId string    `bson:"accountId"`
	MessageId string    `bson:"messageId"`
	Embedding []float32 `bson:"embedding"`
	// useful metadata
	Sender    PersonInfo   `bson:"sender"`
	Receiver  []PersonInfo `bson:"receiver"`
	Summary   string       `bson:"summary"`
	Version   int          `bson:"version"`
	UpdatedAt time.Time    `bson:"updatedAt"`
	CreatedAt time.Time    `bson:"createdAt"`
}

func (g EmailSummaryEmbedding) ToDocumentId() string {
	return ToDocumentId(g.AccountId, g.MessageId)
}
