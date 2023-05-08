package sender

import (
	"bufio"
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

// func main() {
// 	// 调用 readBlacklist 函数读取数据
// 	m, err := readBlacklist()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}

// 	// 输出 map
// 	fmt.Println(m)
// }
