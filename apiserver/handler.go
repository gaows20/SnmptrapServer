package apiserver

import (
	"cqrcsnmpserver/trap"
	"cqrcsnmpserver/common/sender"
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

func handlerIndex(w http.ResponseWriter, r *http.Request){
	htmlByte,err := ioutil.ReadFile("./webapp/index.html")
	if err != nil {
		log.Fatal(fmt.Sprintf("read index.html file error:[%s]", err))
	}
	t,err := template.New("index").Funcs(template.FuncMap{"Replace": Replace}).Parse(string(htmlByte))
	if err != nil {
		log.Fatal(fmt.Sprintf("[parse index.html error:[%s]", err))
	}

	data := map[string][]template.HTML{}
	for k, _ := range trap.TrapMap{
		arr, err := trap.TrapMap[k].GetListArray()
		if err != nil {
			log.WithField("ipaddr", k).Error(fmt.Sprintf("get arraylist from trapmap error", err))
			continue
		}
		traparr := make([]template.HTML,0,0)
		for _, v := range arr{
			if v != nil {
				msg :=  strings.Replace(v.(string),"\n", "<br/>", -1)
				traparr = append(traparr, template.HTML(msg))
			}

		}
		data[k] = traparr
	}
	t.Execute(w, data) //第二个参数表示向模版传递的数
	return
}



func handlerPduDel(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	ip := r.Form.Get("ip")
	index, err := strconv.ParseInt(r.Form.Get("index"),10,64)
	if err != nil {
		fmt.Fprint(w, fmt.Sprintf("index数据[%v]不为int类型，",r.Form.Get("index")))
	}
	if err := trap.DelItem(ip, index); err != nil {
		fmt.Fprint(w, fmt.Sprintf("删除数据失败：%s", err))
	}
	sender.PushRecoverMetrics(ip)
	fmt.Fprint(w, "删除数据成功")


}

