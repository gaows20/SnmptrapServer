package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MibEntry 表示一个 MIB 条目
type MibEntry struct {
	Name        string
	OID         string
	Description string
}

// MibParser 用于解析 MIB 文件
type MibParser struct {
	oidMap      map[string]string // 名称到 OID 的映射
	parentMap   map[string]string // 名称到父 OID 的映射
	entries     []MibEntry        // 解析出的条目
	currentFile string            // 当前正在解析的文件
}

// NewMibParser 创建新的 MIB 解析器
func NewMibParser() *MibParser {
	parser := &MibParser{
		oidMap:    make(map[string]string),
		parentMap: make(map[string]string),
		entries:   make([]MibEntry, 0),
	}

	// 初始化一些已知的基础 OID（H3C 企业 OID）
	// hh3cCommon 通常映射到 1.3.6.1.4.1.25506
	parser.oidMap["hh3cCommon"] = "1.3.6.1.4.1.25506"

	return parser
}

// parseOIDAssignment 解析 OID 赋值语句，如 "name ::= { parent number }"
func (p *MibParser) parseOIDAssignment(line string) {
	// 匹配 OBJECT IDENTIFIER 定义: name OBJECT IDENTIFIER ::= { parent number }
	pattern1 := regexp.MustCompile(`(\w+)\s+OBJECT\s+IDENTIFIER\s*::=\s*\{\s*(\w+)\s+(\d+)\s*\}`)
	matches := pattern1.FindStringSubmatch(line)
	if len(matches) == 4 {
		name := matches[1]
		parent := matches[2]
		number := matches[3]
		p.parentMap[name] = parent + " " + number
		return
	}

	// 匹配简化的 OID 定义: name OBJECT IDENTIFIER ::= { parent }
	pattern2 := regexp.MustCompile(`(\w+)\s+OBJECT\s+IDENTIFIER\s*::=\s*\{\s*(\w+)\s*\}`)
	matches = pattern2.FindStringSubmatch(line)
	if len(matches) == 3 {
		name := matches[1]
		parent := matches[2]
		p.parentMap[name] = parent
		return
	}
}

// parseModuleIdentity 解析 MODULE-IDENTITY 的 OID 定义
func (p *MibParser) parseModuleIdentity(lines []string, startIdx int) (string, int) {
	var moduleName string

	// 提取模块名称
	modulePattern := regexp.MustCompile(`(\w+)\s+MODULE-IDENTITY`)
	if matches := modulePattern.FindStringSubmatch(lines[startIdx]); len(matches) == 2 {
		moduleName = matches[1]
	} else {
		return "", startIdx
	}

	// 查找 ::= { parent number } 行
	for i := startIdx + 1; i < len(lines) && i < startIdx+50; i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "::=") && strings.Contains(line, "{") {
			oidPattern := regexp.MustCompile(`::=\s*\{\s*(\w+)\s+(\d+)\s*\}`)
			if matches := oidPattern.FindStringSubmatch(line); len(matches) == 3 {
				parentName := matches[1]
				number := matches[2]
				p.parentMap[moduleName] = parentName + " " + number
				return moduleName, i
			}
		}
		// 如果遇到下一个主要定义，停止搜索
		if i > startIdx+5 && (strings.Contains(line, "OBJECT-TYPE") ||
			strings.Contains(line, "NOTIFICATION-TYPE") ||
			strings.Contains(line, "OBJECT IDENTIFIER")) {
			break
		}
	}

	return moduleName, startIdx
}

// resolveOID 解析完整的 OID 路径
func (p *MibParser) resolveOID(name string, suffix string) string {
	// 如果已经有缓存的 OID，直接返回
	if oid, ok := p.oidMap[name]; ok {
		if suffix != "" {
			return oid + "." + suffix
		}
		return oid
	}

	// 查找父节点
	if parentInfo, ok := p.parentMap[name]; ok {
		parts := strings.Fields(parentInfo)
		if len(parts) == 2 {
			parentName := parts[0]
			number := parts[1]
			parentOID := p.resolveOID(parentName, "")
			if parentOID != "" {
				oid := parentOID + "." + number
				p.oidMap[name] = oid
				if suffix != "" {
					return oid + "." + suffix
				}
				return oid
			}
		} else if len(parts) == 1 {
			parentName := parts[0]
			parentOID := p.resolveOID(parentName, "")
			if parentOID != "" {
				p.oidMap[name] = parentOID
				if suffix != "" {
					return parentOID + "." + suffix
				}
				return parentOID
			}
		}
	}

	// 如果无法解析，返回空字符串（该条目将被跳过）
	return ""
}

// parseObjectType 解析 OBJECT-TYPE 或 NOTIFICATION-TYPE 定义
func (p *MibParser) parseObjectType(lines []string, startIdx int) (MibEntry, int) {
	var entry MibEntry
	var name string
	var description strings.Builder
	inDescription := false
	descriptionStarted := false

	// 提取名称和类型
	namePattern := regexp.MustCompile(`(\w+)\s+(OBJECT-TYPE|NOTIFICATION-TYPE)`)
	line := strings.TrimSpace(lines[startIdx])
	if matches := namePattern.FindStringSubmatch(line); len(matches) == 3 {
		name = matches[1]
		entry.Name = name
	}

	// 查找 OID 赋值和 DESCRIPTION
	i := startIdx + 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		originalLine := lines[i]

		// 检查是否是 OID 赋值行: ::= { parent number }
		if strings.Contains(line, "::=") && strings.Contains(line, "{") && !inDescription {
			oidPattern := regexp.MustCompile(`::=\s*\{\s*(\w+)\s+(\d+)\s*\}`)
			if matches := oidPattern.FindStringSubmatch(line); len(matches) == 3 {
				parentName := matches[1]
				oidSuffix := matches[2]
				entry.OID = p.resolveOID(parentName, oidSuffix)
				// OID 赋值通常是定义的结束，但可能还有描述
				if !descriptionStarted {
					// 继续查找描述
				}
			}
		}

		// 检查是否是 DESCRIPTION 开始
		if strings.HasPrefix(line, "DESCRIPTION") {
			descriptionStarted = true
			inDescription = true
			// 尝试提取单行描述
			descPattern := regexp.MustCompile(`DESCRIPTION\s+"([^"]*)"`)
			if matches := descPattern.FindStringSubmatch(line); len(matches) == 2 {
				description.WriteString(matches[1])
				inDescription = false
			} else {
				// 多行描述开始
				descPattern2 := regexp.MustCompile(`DESCRIPTION\s+"([^"]*)`)
				if matches := descPattern2.FindStringSubmatch(line); len(matches) == 2 {
					description.WriteString(matches[1])
				} else if strings.HasPrefix(line, "DESCRIPTION") {
					// DESCRIPTION 关键字单独一行，描述在下一行
					inDescription = true
				}
			}
		} else if inDescription {
			// 继续读取描述内容
			// 检查是否是描述结束（以引号结尾）
			if strings.HasSuffix(line, "\"") && !strings.HasPrefix(line, "\"") {
				// 多行描述结束
				desc := strings.TrimSuffix(line, "\"")
				desc = strings.TrimSpace(desc)
				if desc != "" {
					if description.Len() > 0 {
						description.WriteString(" ")
					}
					description.WriteString(desc)
				}
				inDescription = false
			} else if strings.HasPrefix(line, "\"") && strings.HasSuffix(line, "\"") {
				// 单行描述
				desc := strings.Trim(line, "\"")
				if description.Len() > 0 {
					description.WriteString(" ")
				}
				description.WriteString(desc)
				inDescription = false
			} else if strings.HasPrefix(line, "\"") {
				// 描述开始
				desc := strings.TrimPrefix(line, "\"")
				if description.Len() > 0 {
					description.WriteString(" ")
				}
				description.WriteString(desc)
			} else if !strings.HasPrefix(line, "--") && line != "" {
				// 继续描述内容（跳过注释和空行）
				// 检查是否是下一个定义开始
				nextNamePattern := regexp.MustCompile(`^\s*(\w+)\s+(OBJECT-TYPE|NOTIFICATION-TYPE|OBJECT\s+IDENTIFIER|MODULE-IDENTITY)`)
				if nextNamePattern.MatchString(originalLine) {
					inDescription = false
					break
				}
				if description.Len() > 0 {
					description.WriteString(" ")
				}
				description.WriteString(line)
			}
		}

		// 如果遇到下一个定义开始，停止
		if i > startIdx+1 {
			nextNamePattern := regexp.MustCompile(`^\s*(\w+)\s+(OBJECT-TYPE|NOTIFICATION-TYPE|OBJECT\s+IDENTIFIER|MODULE-IDENTITY)`)
			if nextNamePattern.MatchString(originalLine) && !inDescription {
				break
			}
		}

		// 如果已经找到 OID 且描述已结束，可以提前退出
		if entry.OID != "" && !inDescription && descriptionStarted {
			// 检查下一行是否是新的定义
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				nextNamePattern := regexp.MustCompile(`^\s*(\w+)\s+(OBJECT-TYPE|NOTIFICATION-TYPE|OBJECT\s+IDENTIFIER|MODULE-IDENTITY)`)
				if nextNamePattern.MatchString(nextLine) {
					break
				}
			}
		}

		i++
	}

	entry.Description = strings.TrimSpace(description.String())
	return entry, i
}

// parseMibFile 解析单个 MIB 文件
func (p *MibParser) parseMibFile(filepath string) error {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("读取文件失败 %s: %v", filepath, err)
	}

	p.currentFile = filepath
	lines := regexp.MustCompile(`\r?\n`).Split(string(content), -1)

	// 第一遍：解析所有 OBJECT IDENTIFIER 和 MODULE-IDENTITY 定义，构建 OID 树
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// 处理 OBJECT IDENTIFIER 定义
		if strings.Contains(line, "OBJECT IDENTIFIER") && strings.Contains(line, "::=") {
			p.parseOIDAssignment(line)
		}

		// 处理 MODULE-IDENTITY 的 OID 定义
		if strings.Contains(line, "MODULE-IDENTITY") {
			_, nextIdx := p.parseModuleIdentity(lines, i)
			if nextIdx > i {
				i = nextIdx
			}
		}
	}

	// 第二遍：解析 OBJECT-TYPE 和 NOTIFICATION-TYPE
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "OBJECT-TYPE") || strings.Contains(line, "NOTIFICATION-TYPE") {
			entry, nextIdx := p.parseObjectType(lines, i)
			if entry.Name != "" && entry.OID != "" {
				p.entries = append(p.entries, entry)
			} else if entry.Name != "" && entry.OID == "" {
				// 记录无法解析 OID 的条目（用于调试）
				fileName := p.currentFile
				if idx := strings.LastIndex(fileName, string(os.PathSeparator)); idx >= 0 {
					fileName = fileName[idx+1:]
				}
				log.Printf("警告: 无法解析 %s 的 OID (文件: %s)\n", entry.Name, fileName)
			}
			i = nextIdx - 1
		}
	}

	return nil
}

// parseAllMibFiles 解析指定目录下所有 .mib 文件
func (p *MibParser) parseAllMibFiles(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("读取目录失败: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".mib") {
			filePath := filepath.Join(dirPath, file.Name())
			fmt.Printf("正在解析: %s\n", file.Name())
			if err := p.parseMibFile(filePath); err != nil {
				log.Printf("解析文件 %s 时出错: %v\n", filePath, err)
				continue
			}
		}
	}

	return nil
}

// writeToFile 将解析结果写入文件
func (p *MibParser) writeToFile(outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 写入表头（可选）
	// writer.WriteString("name\toid\tdescription\n")

	// 写入所有条目
	for _, entry := range p.entries {
		// 格式: "oidname"			"oid"			""
		line := fmt.Sprintf("\"%s\"\t\t\"%s\"\t\t\"\"\n", entry.Name, entry.OID)
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}
	}

	fmt.Printf("成功解析 %d 个条目，已写入到 %s\n", len(p.entries), outputPath)
	return nil
}

func main() {
	// 获取当前文件所在目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("获取当前目录失败:", err)
	}

	// MIB 文件目录
	mibDir := filepath.Join(currentDir, "H3C New Style Private MIB")
	if _, err := os.Stat(mibDir); os.IsNotExist(err) {
		log.Fatalf("目录不存在: %s", mibDir)
	}

	// 输出文件路径
	outputFile := filepath.Join(currentDir, "h3c_new_style.txt")

	// 创建解析器并解析所有文件
	parser := NewMibParser()
	if err := parser.parseAllMibFiles(mibDir); err != nil {
		log.Fatal("解析 MIB 文件失败:", err)
	}

	// 写入结果文件
	if err := parser.writeToFile(outputFile); err != nil {
		log.Fatal("写入文件失败:", err)
	}
}
