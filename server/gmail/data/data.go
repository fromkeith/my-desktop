package data

import "time"

type PersonInfo struct {
	Email string `validate:"required"`
	Name  string `validate:"required"`
} // @name PersonInfo

type GmailEntry struct {
	UserId              string                  `validate:"required" bson:"userId"`
	MessageId           string                  `validate:"required" bson:"messageId"`
	ThreadId            string                  `validate:"required" bson:"threadId"`
	Labels              []string                `validate:"required" bson:"labels"`
	Subject             string                  `validate:"required" bson:"subject"`
	Snippet             string                  `validate:"required" bson:"snippet"`
	HistoryId           uint64                  `validate:"required" bson:"historyId"`
	InternalDate        int64                   `validate:"required" bson:"internalDate"`
	Headers             map[string]string       `validate:"required" bson:"headers"`
	Sender              PersonInfo              `validate:"required" bson:"sender"`
	Receiver            []PersonInfo            `validate:"required" bson:"receiver"`
	ReceivedAt          string                  `validate:"required" bson:"receivedAt"`
	ReplyTo             *PersonInfo             `json:",omitempty" bson:"replyTo"`
	AdditionalReceivers map[string][]PersonInfo `validate:"required" bson:"additionalReceivers"`
	// used in database, but not returned via API
	AccountId string `json:"-" bson:"accountId"`
	// For Sync + Conflict Resolution
	UpdatedAt     time.Time `validate:"required" bson:"updatedAt"`
	CreatedAt     time.Time `validate:"required" bson:"createdAt"`
	RevisionCount int64     `validate:"required" bson:"revisionCount"`
	// used for mongo to determine upserts vs conflicts
	LastBatchWriteId string `json:"-" bson:"lastBatchWriteId"`
} // @name GmailEntry

func (g GmailEntry) ToDocumentId() string {
	return ToDocumentId(g.AccountId, g.MessageId)
}

func ToDocumentId(accountId, messageId string) string {
	return accountId + ";" + messageId
}

type GmailEntryBody struct {
	UserId         string   `validate:"required" bson:"userId"`
	MessageId      string   `validate:"required" bson:"messageId"`
	PlainText      string   `bson:"plainText"`
	Html           string   `bson:"html"`
	HasAttachments int      `validate:"required" bson:"hasAttachments"`
	AttachmentIds  []string `bson:"attachmentIds"`
	// used in database, but not returned via API
	AccountId string `json:"-" bson:"accountId"`
	// For Sync + Conflict Resolution
	UpdatedAt     time.Time `validate:"required" bson:"updatedAt"`
	CreatedAt     time.Time `validate:"required" bson:"createdAt"`
	RevisionCount int64     `validate:"required" bson:"revisionCount"`
	// used for mongo to determine upserts vs conflicts
	LastBatchWriteId string `json:"-" bson:"lastBatchWriteId"`
} // @name GmailEntryBody

func (g GmailEntryBody) ToDocumentId() string {
	return g.AccountId + ";" + g.MessageId
}
