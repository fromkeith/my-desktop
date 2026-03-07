package threads

import (
	"fromkeith/my-desktop-server/shared/gmail/data"
	"time"
)

type MessageBasic struct {
	MessageId    string          `validate:"required" json:"messageId" bson:"messageId"`
	InternalDate int64           `validate:"required" json:"internalDate" bson:"internalDate"`
	Sender       data.PersonInfo `json:"sender" bson:"sender"`
	Subject      string          `json:"subject" bson:"subject"`
	Snippet      string          `json:"snippet" bson:"snippet"`
	Labels       []string        `validate:"required" json:"labels" bson:"labels"`
} // @name MessageBasic

type ThreadEntry struct {
	Messages               []MessageBasic `validate:"required" json:"messages" bson:"messages"`
	UpdatedAt              time.Time      `validate:"required" json:"updatedAt" bson:"updatedAt"`
	ThreadId               string         `validate:"required" json:"threadId" bson:"threadId"`
	MostRecentInternalDate int64          `validate:"required" json:"mostRecentInternalDate" bson:"mostRecentInternalDate"`
	Categories             []string       `validate:"required" json:"categories" bson:"categories"`
	Tags                   []string       `validate:"required" json:"tags" bson:"tags"`
} // @name Thread
