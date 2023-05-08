package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func parseMibTxt(filepath string) (map[string]string, error) {
	// 读取文件内容
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	// 定义 map
	var m = make(map[string]string)
	// 定义正则表达式
	oidPattern := regexp.MustCompile(`^\.\d+(\.\d+)*$`)
	namePattern := regexp.MustCompile(`([a-zA-Z][\w-]+)\s+NOTIFICATION-TYPE`)
	// descPattern := regexp.MustCompile(`DESCRIPTION\s+"(?:[^"\\]|\\.)*"`)
	// 分割文件内容，按行处理
	lines := regexp.MustCompile(`\r?\n`).Split(string(content), -1)
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if oidPattern.MatchString(line) {
			// 匹配到 OID
			oid := line
			i++
			for i < len(lines) && !oidPattern.MatchString(lines[i]) {

				// 处理 OID 名称和描述信息
				if namePattern.MatchString(lines[i]) {
					// fmt.Printf("Description: %s\n\n", lines[i])
					name := namePattern.FindStringSubmatch(lines[i])[1]
					// fmt.Printf("OID: %s\nName: %s\n\n", oid, name)
					i++
					k := strings.Trim(name, "\"")
					v := strings.Trim(oid[1:], "\"")
					m[k] = v
					// if i < len(lines) && descPattern.MatchString(lines[i]) {
					// 	// fmt.Printf("Description: %s\n\n", lines[i])
					// 	desc := descPattern.FindStringSubmatch(lines[i])[1]
					// 	// 输出结果
					// 	fmt.Printf("OID: %s\nName: %s\nDescription: %s\n\n", oid, name, desc)
					// }
				}
				i++
			}
			i--
		}
	}
	// 创建文件
	file, err := os.Create("h3c_new_style.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 将map写入文件
	for k, v := range m {
		line := fmt.Sprintf("\"%s\"			\"%s\"\n", k, v)
		_, err := file.WriteString(line)
		if err != nil {
			log.Fatal(err)
		}
	}
	return m, nil
}

func main() {
	// 调用 readBlacklist 函数读取数据
	_, err := parseMibTxt("Quick reference of H3C new style MIB objects description.txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 输出 map
	// fmt.Println(m)
}
