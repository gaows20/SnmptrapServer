package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	url := "http://localhost:8000/post"
	data := []byte(`{"hello": "world"}`)

	//创建http客户端
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error occurred while sending POST request: ", err)
	}

	//设置请求头
	req.Header.Set("Content-Type", "application/json")

	//发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error occurred while sending POST request: ", err)
	}

	defer resp.Body.Close()

	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", resp.Body)
}
