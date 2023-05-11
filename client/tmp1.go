package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// 打开miblist.txt文件
	file, err := os.Open("miblist.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 读取每行内容并追加空格后写回文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line) + "			\" \""
		fmt.Println(line)

		// 将修改后的内容写回文件
		f, err := os.OpenFile("miblist2.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		_, err = fmt.Fprintln(f, line)
		if err != nil {
			panic(err)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
