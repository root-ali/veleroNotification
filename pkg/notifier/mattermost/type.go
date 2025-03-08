package mattermost

import (
	"go.uber.org/zap"
	"time"
)

type MattermostClient struct {
	Url      string
	Token    string
	Channel  string
	UserName string
	Timeout  time.Duration
	logger   *zap.SugaredLogger
}

type MattermostMessage struct {
	Username   string             `json:"username"`
	Status     string             `json:"-"`
	Message    string             `json:"tex:omitempty"`
	Attachment []AttachmentObject `json:"attachments"`
	Color      string             `json:"-"`
	IconEmoji  string             `json:"icon_emoji"`
}

type AttachmentObject struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Color string `json:"color"`
}
