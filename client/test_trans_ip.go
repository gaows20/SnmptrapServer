package main

import (
	"fmt"
	"net"
	"strings"
)

type SNMPIPParser struct {
	// 可以添加配置项，如是否严格验证IP有效性等
	StrictValidation bool
}

func main() {
	// 示例数据
	octetString := []byte{0x64, 0x48, 0x17, 0x66} // "dH\u0017f"

	fmt.Println("基础解析:")
	basicIP := parseOctetStringToIP(octetString)
	fmt.Printf(basicIP)

}

// parseAsIP 尝试解析为IP地址
func parseAsIP(data []byte) string {
	// IPv4: 4字节
	if len(data) == 4 {
		ip := net.IPv4(data[0], data[1], data[2], data[3])
		// 验证IP有效性（排除全零等特殊情况）
		if !ip.IsUnspecified() && !ip.IsMulticast() {
			return ip.String()
		}
	}

	// IPv6: 16字节
	if len(data) == 16 {
		ip := net.IP(data)
		if ip.To4() == nil && !ip.IsUnspecified() && !ip.IsMulticast() {
			return ip.String()
		}
	}

	return ""
}

// parseAsMAC 尝试解析为MAC地址
func parseAsMAC(data []byte) string {
	// MAC地址: 6字节
	if len(data) == 6 {
		// 构建MAC地址格式
		hexParts := make([]string, len(data))
		for i, b := range data {
			hexParts[i] = fmt.Sprintf("%02x", b)
		}
		return strings.Join(hexParts, ":")
	}

	return ""
}

// formatAsHexString 将字节数组格式化为十六进制字符串
func formatAsHexString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	hexParts := make([]string, len(data))
	for i, b := range data {
		hexParts[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(hexParts, ":")
}

// 保持向后兼容的函数
func parseOctetStringToIP(data []byte) string {
	return ParseOctetString(data)
}

func bytesToHexString(data []byte) string {
	return formatAsHexString(data)
}

// ParseOctetString 通用解析函数，自动识别IP和MAC地址，解析失败返回原有内容
func ParseOctetString(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// 尝试解析为IP地址 (IPv4: 4字节, IPv6: 16字节)
	if ip := parseAsIP(data); ip != "" {
		return ip
	}

	// 尝试解析为MAC地址 (6字节)
	if mac := parseAsMAC(data); mac != "" {
		return mac
	}

	// 如果都无法解析，返回原始内容的十六进制表示
	return formatAsHexString(data)
}
