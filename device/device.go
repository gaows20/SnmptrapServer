package device

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

var (
	deviceMap    = make(map[string]string)
	deviceMapMux sync.RWMutex
	mapFile      = "device_map.json"
)

// Init 初始化设备映射
func Init() error {
	return loadDeviceMap()
}

// loadDeviceMap 从文件加载设备映射
func loadDeviceMap() error {
	deviceMapMux.Lock()
	defer deviceMapMux.Unlock()

	data, err := ioutil.ReadFile(mapFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，创建空映射
			deviceMap = make(map[string]string)
			return saveDeviceMap()
		}
		return fmt.Errorf("读取设备映射文件失败: %v", err)
	}

	if err := json.Unmarshal(data, &deviceMap); err != nil {
		return fmt.Errorf("解析设备映射文件失败: %v", err)
	}

	return nil
}

// saveDeviceMap 保存设备映射到文件
func saveDeviceMap() error {
	data, err := json.MarshalIndent(deviceMap, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化设备映射失败: %v", err)
	}

	if err := ioutil.WriteFile(mapFile, data, 0644); err != nil {
		return fmt.Errorf("写入设备映射文件失败: %v", err)
	}

	return nil
}

// GetDeviceName 根据IP获取设备名称
func GetDeviceName(ip string) string {
	deviceMapMux.RLock()
	defer deviceMapMux.RUnlock()

	if name, ok := deviceMap[ip]; ok {
		return name
	}
	return ""
}

// SetDeviceName 设置设备名称
func SetDeviceName(ip, name string) error {
	deviceMapMux.Lock()
	defer deviceMapMux.Unlock()

	deviceMap[ip] = name
	return saveDeviceMap()
}

// DeleteDevice 删除设备映射
func DeleteDevice(ip string) error {
	deviceMapMux.Lock()
	defer deviceMapMux.Unlock()

	delete(deviceMap, ip)
	return saveDeviceMap()
}

// GetAllDevices 获取所有设备映射
func GetAllDevices() map[string]string {
	deviceMapMux.RLock()
	defer deviceMapMux.RUnlock()

	result := make(map[string]string)
	for ip, name := range deviceMap {
		result[ip] = name
	}
	return result
}
