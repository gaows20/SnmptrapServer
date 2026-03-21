package apiserver

import (
	"cqrcsnmpserver/common/sender"
	"cqrcsnmpserver/storage"
	"cqrcsnmpserver/trap"
	"fmt"
	log "github.com/sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func Replace(str string) (res string) {
	return  strings.Replace(str, ".", "_", -1)
}

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	htmlByte, err := ioutil.ReadFile("./webapp/index.html")
	if err != nil {
		log.Fatal(fmt.Sprintf("read index.html file error:[%s]", err))
	}
	t, err := template.New("index").Funcs(template.FuncMap{"Replace": Replace}).Parse(string(htmlByte))
	if err != nil {
		log.Fatal(fmt.Sprintf("[parse index.html error:[%s]", err))
	}

	data := map[string][]template.HTML{}

	// 从持久化存储读取消息
	messages := storage.GetTrapMessages(1000)
	if len(messages) > 0 {
		// 按 IP 分组
		msgByIP := make(map[string][]template.HTML)
		for _, msg := range messages {
			formattedMsg := strings.Replace(msg.Content, "\n", "<br/>", -1)
			msgByIP[msg.HostIP] = append(msgByIP[msg.HostIP], template.HTML(formattedMsg))
		}
		data = msgByIP
	} else {
		// 如果没有持久化数据，回退到内存存储
		for k := range trap.TrapMap {
			arr, err := trap.TrapMap[k].GetListArray()
			if err != nil {
				log.WithField("ipaddr", k).Error(fmt.Sprintf("get arraylist from trapmap error", err))
				continue
			}
			traparr := make([]template.HTML, 0)
			for _, v := range arr {
				if v != nil {
					msg := strings.Replace(v.(string), "\n", "<br/>", -1)
					traparr = append(traparr, template.HTML(msg))
				}
			}
			data[k] = traparr
		}
	}
	t.Execute(w, data)
	return
}



func handlerPduDel(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()
	ip := r.Form.Get("ip")
	index, err := strconv.ParseInt(r.Form.Get("index"),10,64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"success": false, "message": "index数据[%v]不为int类型"}`, r.Form.Get("index"))
		return
	}
	if err := trap.DelItem(ip, index); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"success": false, "message": "删除数据失败：%s"}`, err)
		return
	}
	sender.PushRecoverMetrics(ip)
	fmt.Fprint(w, `{"success": true, "message": "删除数据成功"}`)
}

