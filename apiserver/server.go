package apiserver

import (
	"cqrcsnmpserver/global"
	"net"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

// 任务的http接口
type ApiServer struct {
	httpServer *http.Server
}

var (
	G_apiServer *ApiServer
)

//初始化http服务
func InitAppServer() (err error) {
	var (
		listener net.Listener // apiserver的监听器
	)
	//配置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerIndex)
	mux.HandleFunc("/delpdu", handlerPduDel)

	//设置静态文件存储目录
	staticDir := http.Dir(global.GVA_CONFIG.ApiServer.ApiWebRoot)
	staticHandler := http.FileServer(staticDir)
	mux.Handle("/static/", http.StripPrefix("/", staticHandler))

	//启动TCP监听
	log.WithField("port", global.GVA_CONFIG.ApiServer.ApiPort).Info("start appserver listen on the port")
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa((global.GVA_CONFIG.ApiServer.ApiPort))); err != nil {
		return err
	}

	httpServer := &http.Server{
		ReadHeaderTimeout: time.Duration(global.GVA_CONFIG.ApiServer.ApiReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(global.GVA_CONFIG.ApiServer.ApiWriteTimeout) * time.Second,
		Handler:           mux,
	}

	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	go httpServer.Serve(listener)

	return nil
}
