package main

import (
	"bufio"
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

// compareAndAppendMibFiles 比较 h3c_new_style.txt 和 miblist.txt 两个文件的差异
// 如果 h3c_new_style.txt 中有数据不在 miblist.txt 中，则按照 miblist.txt 的格式补充到 miblist.txt 中
// 参数:
//   h3cFile: h3c_new_style.txt 文件的路径（相对路径或绝对路径）
//   miblistFile: miblist.txt 文件的路径（相对路径或绝对路径）
// 使用示例:
//   err := compareAndAppendMibFiles("h3c_new_style.txt", "../../miblist.txt")
func compareAndAppendMibFiles(h3cFile string, miblistFile string) error {
	// 读取 h3c_new_style.txt 文件
	h3cContent, err := ioutil.ReadFile(h3cFile)
	if err != nil {
		return fmt.Errorf("读取 h3c_new_style.txt 文件失败: %v", err)
	}

	// 读取 miblist.txt 文件
	miblistContent, err := ioutil.ReadFile(miblistFile)
	if err != nil {
		return fmt.Errorf("读取 miblist.txt 文件失败: %v", err)
	}

	// 解析 h3c_new_style.txt，提取 name 和 oid
	// 格式: "name"			"oid"
	h3cMap := make(map[string]string) // key: name, value: oid
	h3cLines := regexp.MustCompile(`\r?\n`).Split(string(h3cContent), -1)
	for _, line := range h3cLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 使用正则表达式匹配引号内的内容，更健壮地处理制表符和空格
		re := regexp.MustCompile(`"([^"]+)"\s+"([^"]+)"`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			name := matches[1]
			oid := matches[2]
			if name != "" && oid != "" {
				h3cMap[name] = oid
			}
		}
	}

	// 解析 miblist.txt，提取已存在的 name 和 oid
	// 格式: "name"			"oid"			"description"
	existingMap := make(map[string]bool) // key: name+"\t"+oid, value: true
	miblistLines := regexp.MustCompile(`\r?\n`).Split(string(miblistContent), -1)
	for _, line := range miblistLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 使用正则表达式匹配前两个引号内的内容
		re := regexp.MustCompile(`"([^"]+)"\s+"([^"]+)"`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			name := matches[1]
			oid := matches[2]
			if name != "" && oid != "" {
				// 使用 name 和 oid 作为唯一标识
				key := name + "\t" + oid
				existingMap[key] = true
			}
		}
	}

	// 找出需要补充的数据
	var toAppend []string
	for name, oid := range h3cMap {
		key := name + "\t" + oid
		if !existingMap[key] {
			// 按照 miblist.txt 的格式：三列，第三列为空字符串
			line := fmt.Sprintf("\"%s\"\t\t\"%s\"\t\t\"\"", name, oid)
			toAppend = append(toAppend, line)
		}
	}

	// 如果有需要补充的数据，追加到 miblist.txt
	if len(toAppend) > 0 {
		// 以追加模式打开文件
		file, err := os.OpenFile(miblistFile, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("打开 miblist.txt 文件失败: %v", err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		// 先写入一个换行，确保新内容从新行开始
		_, err = writer.WriteString("\n")
		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		// 写入需要补充的数据
		for _, line := range toAppend {
			_, err = writer.WriteString(line + "\n")
			if err != nil {
				return fmt.Errorf("写入文件失败: %v", err)
			}
		}

		err = writer.Flush()
		if err != nil {
			return fmt.Errorf("刷新文件缓冲区失败: %v", err)
		}

		fmt.Printf("成功补充 %d 条数据到 miblist.txt\n", len(toAppend))
	} else {
		fmt.Println("没有需要补充的数据")
	}

	return nil
}

func main() {
	// 调用 readBlacklist 函数读取数据
	// _, err := parseMibTxt("Quick reference of H3C new style MIB objects description.txt")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// 输出 map
	// fmt.Println(m)
	err := compareAndAppendMibFiles("h3c_new_style.txt", "../../miblist.txt")
	if err != nil {
		fmt.Println(err)
	}
}
