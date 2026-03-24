package main

import (
	"cqrcsnmpserver/apiserver"
	"cqrcsnmpserver/core"
	"cqrcsnmpserver/device"
	"cqrcsnmpserver/global"
	"cqrcsnmpserver/trap"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	global.GVA_VP = core.Viper()
	core.InitLog()

	// 初始化设备映射
	if err := device.Init(); err != nil {
		log.WithError(err).Error("初始化设备映射失败")
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