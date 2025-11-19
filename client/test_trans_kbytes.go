package main

import (
	"fmt"
	"math"
	"strconv"
)

// 定义处理函数类型（现在接收字符串参数）
type ValueHandler func(string) string

// 全局value map - 简洁版本
var ValueMap map[string]ValueHandler = map[string]ValueHandler{
	"hwCurrentStatisticalPeriodRate": convertRateSimple,
	"hwLastStatisticalPeriodRate":    convertRateSimple,
}

// convertRateSimple 简单速率转换（直接向上取整）
func convertRateSimple(kbpsStr string) string {
	// 将字符串转换为整数
	kbps, err := strconv.ParseFloat(kbpsStr, 64)
	if err != nil {
		// 如果转换失败，返回原始字符串
		return kbpsStr
	}

	switch {
	case kbps >= 1000000: // 转换为Gbps
		gbps := kbps / 1000000
		return fmt.Sprintf("%.0fGbps", math.Ceil(gbps))
	case kbps >= 1000: // 转换为Mbps
		mbps := kbps / 1000
		return fmt.Sprintf("%.0fMbps", math.Ceil(mbps))
	default: // 保持kbps
		return fmt.Sprintf("%.0fkbps", math.Ceil(kbps))
	}
}

// 处理SNMP陷阱数据
func ProcessSNMPTrap(trapData map[string]string) map[string]string {
	results := make(map[string]string)

	for oid, value := range trapData {
		if handler, exists := ValueMap[oid]; exists {
			results[oid] = handler(value)
		} else {
			results[oid] = value // 直接返回原始字符串值
		}
	}

	return results
}

func main() {
	// 模拟SNMP陷阱数据（字符串格式的整数）
	trapData := map[string]string{
		"hwCurrentStatisticalPeriodRate": "34255862", // 2.5Gbps -> 3Gbps
		"hwLastStatisticalPeriodRate":    "14112466", // 1.52Mbps -> 2Mbps
	}

	fmt.Println("处理SNMP陷阱数据:")
	results := ProcessSNMPTrap(trapData)

	for oid, result := range results {
		fmt.Printf("  %s: %s -> %s\n", oid, trapData[oid], result)
	}
}
