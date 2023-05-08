package sender

import (
	"cqrcsnmpserver/global"
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

var black_mib_tree map[string]string

var black_list_error error
var once sync.Once
var senderFunc map[string]func(string, global.PushMessage) error = make(map[string]func(string, global.PushMessage) error)

func init() {
	// senderFunc["zabbix"] = SendToZabbix
	senderFunc["webhook"] = PushWebhooks
	senderFunc["pushgateway"] = PushMetrics
}

func Sends(host string, msgheader global.PushMessage) {

	once.Do(func() {
		black_mib_tree, black_list_error = ReadBlacklist(global.GVA_CONFIG.TrapServer.BlackMibMapFile)
		if black_list_error != nil {
			fmt.Println(black_list_error)
			return
		}
	})
	if len(black_mib_tree) == 0 {
		fmt.Println("black_mib_tree is empty 黑名单是空的")
	}
	// 判断是否在黑名单里
	for _, item := range msgheader.Message {
		parts := strings.Split(item["oid"], ".")
		_, ok := black_mib_tree[parts[0]]
		if ok {
			fmt.Println(parts, "存在于 black_mib_tree 中")
			return
		} else {
			fmt.Println(parts, "不存在于 black_mib_tree 中")
		}
		if item["type"] == "ObjectIdentifier" {
			parts := strings.Split(item["value"], ".")
			_, ok := black_mib_tree[parts[0]]
			if ok {
				fmt.Println(parts, "存在于 black_mib_tree 中")
				return
			} else {
				fmt.Println(parts, "不存在于 black_mib_tree 中")
			}
		}
	}

	for _, v := range global.GVA_CONFIG.Sender.Senders {
		if _, ok := senderFunc[v]; ok {
			if err := senderFunc[v](host, msgheader); err != nil {
				log.WithField("err", err).Error("send trap info error")
			}
		} else {
			log.Error(fmt.Sprintf("no such send method:[%s]", v))
		}
	}
}
