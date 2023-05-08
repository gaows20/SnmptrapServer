package global

import (
	"cqrcsnmpserver/config"

	"github.com/spf13/viper"
)

type PushMessage struct {
	Message   []map[string]string `json:"message"`
	Host      string              `json:"host"`
	Version   string              `json:"version"`
	Status    string              `json:"status"`
	MessageID string              `json:"message_id"`
	Index     string              `json:"index"`
}

var (
	GVA_VP     *viper.Viper
	GVA_CONFIG *config.Server
)
