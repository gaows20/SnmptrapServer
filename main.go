package main

import (
	"cqrcsnmpserver/apiserver"
	"cqrcsnmpserver/core"
	"cqrcsnmpserver/global"
	"cqrcsnmpserver/storage"
	"cqrcsnmpserver/trap"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	global.GVA_VP = core.Viper()
	core.InitLog()

	// 初始化持久化存储
	dataDir := "./data"
	if global.GVA_CONFIG.TrapServer.DataDir != "" {
		dataDir = global.GVA_CONFIG.TrapServer.DataDir
	}
	maxMessages := 10000
	if global.GVA_CONFIG.TrapServer.MaxMessages > 0 {
		maxMessages = global.GVA_CONFIG.TrapServer.MaxMessages
	}

	if _, err := storage.InitStorage(dataDir, maxMessages); err != nil {
		log.WithError(err).Error("初始化存储失败，将使用内存存储")
	} else {
		log.Info("持久化存储初始化成功")
	}

	// 启动api http服务
	log.Info("start running api server")
	if err := apiserver.InitAppServer(); err != nil {
		log.Fatalf("初始化API Server报错%s", err)
		os.Exit(1)
	}
	if trapserver, err := trap.NewTrapServer("0.0.0.0", "162"); err != nil {
		log.WithField("err", err).Fatalf("config TrapServer err")
	} else {
		log.Info("start running trap server")
		trapserver.Run()
	}
}