package gmail_client

type PersonInfo struct {
	Email string `validate:"required"`
	Name  string `validate:"required"`
} // @name PersonInfo

type GmailEntry struct {
	UserId              string                  `validate:"required"`
	MessageId           string                  `validate:"required"`
	ThreadId            string                  `validate:"required"`
	Labels              []string                `validate:"required"`
	Subject             string                  `validate:"required"`
	Snippet             string                  `validate:"required"`
	HistoryId           uint64                  `validate:"required"`
	InternalDate        int64                   `validate:"required"`
	Headers             map[string]string       `validate:"required"`
	Sender              PersonInfo              `validate:"required"`
	Receiver            []PersonInfo            `validate:"required"`
	ReceivedAt          string                  `validate:"required"`
	ReplyTo             *PersonInfo             `json:",omitempty"`
	AdditionalReceivers map[string][]PersonInfo `validate:"required"`
} // @name GmailEntry

type GmailEntryBody struct {
	UserId         string `validate:"required"`
	MessageId      string `validate:"required"`
	PlainText      string
	Html           string
	HasAttachments int `validate:"required"`
	AttachmentIds  []string
} // @name GmailEntryBody
