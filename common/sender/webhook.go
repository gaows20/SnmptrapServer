package sender

import (
	"bytes"
	"cqrcsnmpserver/global"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Message struct {
	Message []map[string]string `json:"message"`
	Host    string              `json:"host"`
}

func PushWebhooks(host string, msg global.PushMessage, msg_info string) error {
	jsonStr, err := json.Marshal(msg)
	if err != nil {
		log.WithError(err).Error("序列化 webhook 消息失败")
		return err
	}

	url := global.GVA_CONFIG.Sender.WebhookUrl
	log.WithFields(log.Fields{
		"url":  url,
		"host": host,
	}).Info("发送 webhook 推送")

	client := &http.Client{Timeout: time.Second * 10}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.WithError(err).Error("创建 webhook 请求失败")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.WithError(err).Error("webhook 推送请求失败")
		return err
	}
	defer resp.Body.Close()

	log.WithFields(log.Fields{
		"status_code": resp.StatusCode,
		"url":         url,
	}).Info("webhook 推送完成")

	return nil
}
