package data

// input to email-injestor. Kafka topic:email_injest
type EmailInjestPayload struct {
	MessageId string
	AccountId string
	UserId    string
}

// output of email-injestor. Kafka topic:email_injest_available
type EmailInjestedPayload struct {
	MessageId string
	AccountId string
	Entry     GmailEntry
}
