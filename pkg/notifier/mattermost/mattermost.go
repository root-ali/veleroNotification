package mattermost

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func NewMattermostClient(url, token, channel string, time time.Duration, logger *zap.SugaredLogger) *MattermostClient {
	return &MattermostClient{
		Url:     url,
		Token:   token,
		Channel: channel,
		Timeout: time,
		logger:  logger,
	}
}

func (mc *MattermostClient) SendMessage(message, status string) error {
	ms := MattermostMessage{}
	ms.Username = mc.UserName

	if status == "Failed" {
		ms.Message = message
		ms.Color = "#FF0000"
		ms.IconEmoji = ":firecracker:"
		ms.Status = status
	} else if status == "Success" {
		ms.Status = status
		ms.Message = message
		ms.IconEmoji = ":large_green_square:"
		ms.Color = "#008000"
	}
	ms.Attachment = []AttachmentObject{
		{
			Title: "Velero Backup" + " " + status + " " + ms.IconEmoji,
			Text:  ms.Message,
			Color: ms.Color,
		},
	}
	url := mc.Url + "/hooks/" + mc.Token
	client := &http.Client{Timeout: mc.Timeout}
	reqBody, err := json.Marshal(ms)
	if err != nil {
		mc.logger.Errorw("Cannot serialize to json", "body", ms)
		return err
	}

	mc.logger.Infow("about to call mattermost endpoint", "message", message, "status", status)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		mc.logger.Errorw("Error creating request", "error", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		mc.logger.Errorw("Error in sending request", "error", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		mc.logger.Errorw("Message not send", "statusCode", resp.StatusCode, "Error Body is", resp.Body)
		return err
	}
	mc.logger.Info("Successfully send message to mattermost", "message", message, "status", status)
	return nil
}
