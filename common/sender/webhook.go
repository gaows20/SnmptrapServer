package sender

import (
	"bytes"
	"cqrcsnmpserver/global"
	"encoding/json"
	"fmt"
	"net/http"
)

type Message struct {
	Message []map[string]string `json:"message"`
	Host    string              `json:"host"`
}

func PushWebhooks(host string, msg global.PushMessage, msg_info string) (err error) {
	// fmt.Println(msg)
	jsonStr, err := json.Marshal(msg)
	url := global.GVA_CONFIG.Sender.WebhookUrl
	//创建http客户端
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println("Error occurred while sending POST request: ", err)
	}
	//设置请求头
	req.Header.Set("Content-Type", "application/json")
	//发送请求
	if resp, err := client.Do(req); err != nil {
		defer resp.Body.Close()
		return err
	}
	return nil
}
