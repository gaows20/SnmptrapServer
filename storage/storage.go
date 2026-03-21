package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TrapMessage 表示存储的 Trap 消息结构
type TrapMessage struct {
	ID        int64     `json:"id"`
	HostIP    string    `json:"host_ip"`
	Version   string    `json:"version"`
	Community string    `json:"community"`
	TrapOID   string    `json:"trap_oid"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Storage 消息存储管理器
type Storage struct {
	mu       sync.RWMutex
	messages []TrapMessage
	dataDir  string
	maxSize  int
	nextID   int64
}

var (
	defaultStorage *Storage
	once           sync.Once
)

// InitStorage 初始化存储
func InitStorage(dataDir string, maxSize int) (*Storage, error) {
	var err error
	once.Do(func() {
		defaultStorage, err = NewStorage(dataDir, maxSize)
	})
	return defaultStorage, err
}

// NewStorage 创建新的存储实例
func NewStorage(dataDir string, maxSize int) (*Storage, error) {
	if maxSize <= 0 {
		maxSize = 10000 // 默认最大存储 10000 条
	}

	// 创建数据目录
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %v", err)
	}

	s := &Storage{
		messages: make([]TrapMessage, 0),
		dataDir:  dataDir,
		maxSize:  maxSize,
		nextID:   1,
	}

	// 加载已有数据
	if err := s.load(); err != nil {
		logrus.WithError(err).Warn("加载历史数据失败，将创建新的存储")
	}

	// 启动定期保存协程
	go s.periodicSave()

	return s, nil
}

// getDataFilePath 获取数据文件路径
func (s *Storage) getDataFilePath() string {
	return filepath.Join(s.dataDir, "trap_messages.json")
}

// load 从文件加载数据
func (s *Storage) load() error {
	dataFile := s.getDataFilePath()

	// 检查文件是否存在
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return nil // 文件不存在，返回空
	}

	data, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("读取数据文件失败: %v", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.Unmarshal(data, &s.messages); err != nil {
		return fmt.Errorf("解析数据文件失败: %v", err)
	}

	// 更新 nextID
	for _, msg := range s.messages {
		if msg.ID >= s.nextID {
			s.nextID = msg.ID + 1
		}
	}

	logrus.WithField("count", len(s.messages)).Info("加载历史消息完成")
	return nil
}

// save 保存数据到文件
func (s *Storage) save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.messages, "", "  ")
	s.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	dataFile := s.getDataFilePath()
	tmpFile := dataFile + ".tmp"

	// 写入临时文件
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %v", err)
	}

	// 原子重命名
	if err := os.Rename(tmpFile, dataFile); err != nil {
		return fmt.Errorf("重命名文件失败: %v", err)
	}

	return nil
}

// periodicSave 定期保存数据
func (s *Storage) periodicSave() {
	ticker := time.NewTicker(30 * time.Second) // 每 30 秒保存一次
	defer ticker.Stop()

	for range ticker.C {
		if err := s.save(); err != nil {
			logrus.WithError(err).Error("定期保存数据失败")
		}
	}
}

// SaveTrapMessage 保存 Trap 消息
func (s *Storage) SaveTrapMessage(hostIP, version, community, trapOID, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := TrapMessage{
		ID:        s.nextID,
		HostIP:    hostIP,
		Version:   version,
		Community: community,
		TrapOID:   trapOID,
		Content:   content,
		CreatedAt: time.Now(),
	}

	s.nextID++

	// 插入到头部（最新的在前面）
	s.messages = append([]TrapMessage{msg}, s.messages...)

	// 限制大小
	if len(s.messages) > s.maxSize {
		s.messages = s.messages[:s.maxSize]
	}

	return nil
}

// GetTrapMessages 获取消息列表
func (s *Storage) GetTrapMessages(limit int) []TrapMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.messages) {
		limit = len(s.messages)
	}

	result := make([]TrapMessage, limit)
	copy(result, s.messages[:limit])
	return result
}

// GetTrapMessagesByHost 根据主机 IP 获取消息
func (s *Storage) GetTrapMessagesByHost(hostIP string, limit int) []TrapMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []TrapMessage
	for _, msg := range s.messages {
		if msg.HostIP == hostIP {
			result = append(result, msg)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result
}

// DeleteOldMessages 删除指定天数之前的消息
func (s *Storage) DeleteOldMessages(days int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	var newMessages []TrapMessage

	for _, msg := range s.messages {
		if msg.CreatedAt.After(cutoff) {
			newMessages = append(newMessages, msg)
		}
	}

	deleted := len(s.messages) - len(newMessages)
	s.messages = newMessages

	logrus.WithField("deleted", deleted).Info("清理旧消息完成")
	return s.save()
}

// Close 关闭存储，保存数据
func (s *Storage) Close() error {
	return s.save()
}

// ==================== 全局函数 ====================

// SaveTrapMessage 保存 Trap 消息（使用默认存储）
func SaveTrapMessage(hostIP, version, community, trapOID, content string) error {
	if defaultStorage == nil {
		return fmt.Errorf("存储未初始化")
	}
	return defaultStorage.SaveTrapMessage(hostIP, version, community, trapOID, content)
}

// GetTrapMessages 获取消息列表（使用默认存储）
func GetTrapMessages(limit int) []TrapMessage {
	if defaultStorage == nil {
		return nil
	}
	return defaultStorage.GetTrapMessages(limit)
}

// GetTrapMessagesByHost 根据主机 IP 获取消息（使用默认存储）
func GetTrapMessagesByHost(hostIP string, limit int) []TrapMessage {
	if defaultStorage == nil {
		return nil
	}
	return defaultStorage.GetTrapMessagesByHost(hostIP, limit)
}

// DeleteOldMessages 删除旧消息（使用默认存储）
func DeleteOldMessages(days int) error {
	if defaultStorage == nil {
		return fmt.Errorf("存储未初始化")
	}
	return defaultStorage.DeleteOldMessages(days)
}

// Close 关闭默认存储
func Close() error {
	if defaultStorage == nil {
		return nil
	}
	return defaultStorage.Close()
}
