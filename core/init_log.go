package core

import (
	"cqrcsnmpserver/global"
	"cqrcsnmpserver/utils"
	"io"
	"path/filepath"
	"fmt"
	log "github.com/sirupsen/logrus"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"os"
	"time"
)
var level log.Level

func InitLog() {
	if ok, _ := utils.PathExists(global.GVA_CONFIG.LogConf.Director); !ok { // 判断是否有Director文件夹
		fmt.Printf("创建保存日志的文件夹，目录为：%v\n", global.GVA_CONFIG.LogConf.Director)
		_ = os.Mkdir(global.GVA_CONFIG.LogConf.Director, os.ModePerm)
	}
	switch global.GVA_CONFIG.LogConf.Level { //初始化配置文件的Level
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "dpanic":
		level = log.PanicLevel
	case "panic":
		level = log.PanicLevel
	case "fatal":
		level = log.FatalLevel
	default:
		level = log.InfoLevel
	}
	if global.GVA_CONFIG.LogConf.Format == "json"{
		log.SetFormatter(&log.JSONFormatter{})
	} else if global.GVA_CONFIG.LogConf.Format == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{})
	}
	current_path, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取程序当前目录失败，%s", err)
		os.Exit(2)
	}
	last_path := filepath.Join(current_path, global.GVA_CONFIG.LogConf.LinkName)
	writer, _ := rotatelogs.New(
		filepath.Join(current_path,global.GVA_CONFIG.LogConf.Director,"snmpserver")+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(last_path),
		rotatelogs.WithMaxAge(time.Duration(180)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(60)*time.Second),
	)
	log.SetOutput(os.Stdout)
	writers := []io.Writer{writer, os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)
	if err == nil {
		log.SetOutput(fileAndStdoutWriter)
	} else {
		fmt.Printf("添加日志文件失败，%s", err)
	}
	log.SetLevel(level)
	log.SetReportCaller(true)
}
