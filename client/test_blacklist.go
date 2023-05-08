package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadBlacklist(filepath string) (map[string]string, error) {
	// 打开 blacklist.txt 文件
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 定义 map
	var m = make(map[string]string)

	// 使用 bufio 读取文件内容，按行进行分割
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// 按空格分割每行数据
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			// 将第一个元素作为 map 的键，第二个元素作为 map 的值
			// m[fields[0]] = fields[1]
			k := strings.Trim(fields[0], "\"")
			v := strings.Trim(fields[1], "\"")
			m[k] = v
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return m, nil
}
func main() {
	black_mib_tree, _ := ReadBlacklist("../blackmiblist.txt")
	// black_mib_tree := map[string]string{
	// 	"hh3cPeriodicalTrap":           "1.3.6.1.4.1.25506.2.38.1.6.3.0.1",
	// 	"syslogMsgAppName":             "1.3.6.1.2.1.192.1.2.1.7",
	// 	"syslogMsgEnableNotifications": "1.3.6.1.2.1.192.1.1.2",
	// 	"syslogMsgFacility":            "1.3.6.1.2.1.192.1.2.1.2",
	// }
	for k, v := range black_mib_tree {
		fmt.Printf("%q", k)
		fmt.Print(strings.TrimSpace(k) == "hh3cPeriodicalTrap")
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Printf("black_mib_tree 的类型是 %T\n", black_mib_tree)
	// fmt.Println(black_mib_tree)
	_, ok := black_mib_tree["hh3cPeriodicalTrap"]
	if ok {
		fmt.Println("hh3cPeriodicalTrap 存在于 black_mib_tree 中")
	} else {
		fmt.Println("hh3cPeriodicalTrap 不存在于 black_mib_tree 中")
	}
}
