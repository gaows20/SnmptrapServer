package apiserver

import (
	"cqrcsnmpserver/common/sender"
	"cqrcsnmpserver/global"
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
	// 发送恢复通知
	recoverMsg := global.PushMessage{
		Host:       ip,
		TrapStatus: 0,
	}
	sender.PushRecoverMetrics(ip)
	sender.PushWebhooks(ip, recoverMsg, "")
	fmt.Fprint(w, `{"success": true, "message": "删除数据成功"}`)
}

func handlerPduBatchDel(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()
	ip := r.Form.Get("ip")
	indicesStr := r.Form.Get("indices")
	
	if ip == "" || indicesStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"success": false, "message": "IP和indices参数不能为空"}`)
		return
	}
	
	indices := strings.Split(indicesStr, ",")
	successCount := 0
	
	for _, indexStr := range indices {
		index, err := strconv.ParseInt(strings.TrimSpace(indexStr), 10, 64)
		if err != nil {
			continue
		}
		if err := trap.DelItem(ip, index); err != nil {
			log.WithError(err).WithFields(log.Fields{
				"ip":    ip,
				"index": index,
			}).Error("批量删除消息失败")
			continue
		}
		successCount++
	}
	
	// 发送恢复通知
	recoverMsg := global.PushMessage{
		Host:       ip,
		TrapStatus: 0,
	}
	sender.PushRecoverMetrics(ip)
	sender.PushWebhooks(ip, recoverMsg, "")
	
	fmt.Fprintf(w, `{"success": true, "message": "成功删除 %d 条消息"}`, successCount)
}

