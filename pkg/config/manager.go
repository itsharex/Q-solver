package config

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"Q-Solver/pkg/logger"
)

// ConfigManager 统一管理配置的加载、验证、持久化和订阅
type ConfigManager struct {
	config      Config
	mu          sync.RWMutex
	configPath  string
	subscribers []func(Config)
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	cm := &ConfigManager{
		config:      NewDefaultConfig(),
		subscribers: make([]func(Config), 0),
	}

	// 设置配置文件路径
	cm.configPath = cm.getConfigPath()

	return cm
}

// getConfigPath 获取配置文件路径
func (cm *ConfigManager) getConfigPath() string {
	configDir := "."
	appDir := filepath.Join(configDir, "config")
	_ = os.MkdirAll(appDir, 0755)

	return filepath.Join(appDir, "config.json")
}

// Load 加载配置，优先级：文件 > 环境变量 > 默认值
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 1. 从默认值开始
	cm.config = NewDefaultConfig()

	// 2. 从环境变量覆盖敏感配置
	cm.loadFromEnv()

	// 3. 从文件加载（如果存在）
	if err := cm.loadFromFile(); err != nil {
		logger.Printf("加载配置文件失败 (使用默认配置): %v", err)
	}

	logger.Println("配置已加载")
	return nil
}

// loadFromEnv 从环境变量加载敏感配置
func (cm *ConfigManager) loadFromEnv() {
	if apiKey := os.Getenv("GHOST_API_KEY"); apiKey != "" {
		cm.config.APIKey = apiKey
	}
	if baseURL := os.Getenv("GHOST_BASE_URL"); baseURL != "" {
		cm.config.BaseURL = baseURL
	}
}

// loadFromFile 从文件加载配置
func (cm *ConfigManager) loadFromFile() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在不是错误
		}
		return err
	}

	// 解析到临时结构，避免覆盖未设置的字段
	var fileConfig Config
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 合并配置（文件配置覆盖默认值，但不覆盖环境变量设置的敏感信息）
	cm.mergeConfig(&fileConfig)

	return nil
}

// mergeConfig 合并配置
func (cm *ConfigManager) mergeConfig(src *Config) {
	// 只合并非空/非默认值的字段
	if src.APIKey != "" && os.Getenv("GHOST_API_KEY") == "" {
		cm.config.APIKey = src.APIKey
	}
	if src.BaseURL != "" && os.Getenv("GHOST_BASE_URL") == "" {
		cm.config.BaseURL = src.BaseURL
	}
	if src.Model != "" {
		cm.config.Model = src.Model
	}
	if src.Prompt != "" {
		cm.config.Prompt = src.Prompt
	}
	if src.Opacity != 0 {
		cm.config.Opacity = src.Opacity
	}
	if src.ScreenshotMode != "" {
		cm.config.ScreenshotMode = src.ScreenshotMode
	}
	if src.CompressionQuality != 0 {
		cm.config.CompressionQuality = src.CompressionQuality
	}
	if src.Sharpening != 0 {
		cm.config.Sharpening = src.Sharpening
	}
	if src.ResumePath != "" {
		cm.config.ResumePath = src.ResumePath
	}
	if src.ResumeContent != "" {
		cm.config.ResumeContent = src.ResumeContent
	}

	// 布尔值需要特殊处理（因为 false 是零值）
	cm.config.Grayscale = src.Grayscale
	cm.config.NoCompression = src.NoCompression
	cm.config.KeepContext = src.KeepContext
	cm.config.InterruptThinking = src.InterruptThinking
	cm.config.UseMarkdownResume = src.UseMarkdownResume
	cm.config.UseSpeechToText = src.UseSpeechToText
	cm.config.VoiceListening = src.VoiceListening

	// 合并快捷键
	if len(src.Shortcuts) > 0 {
		maps.Copy(cm.config.Shortcuts, src.Shortcuts)
	}
}

// Save 保存配置到文件
func (cm *ConfigManager) Save() error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	logger.Printf("配置已保存到: %s", cm.configPath)
	return nil
}

// Get 获取当前配置的只读副本
func (cm *ConfigManager) Get() Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config
}

// GetPtr 获取配置指针（用于需要引用的场景，慎用）
func (cm *ConfigManager) GetPtr() *Config {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return &cm.config
}

// Update 更新配置（部分更新）
func (cm *ConfigManager) Update(patch ConfigPatch) error {
	cm.mu.Lock()

	// 应用补丁
	cm.applyPatch(patch)

	// 复制一份用于通知（避免在锁内调用回调）
	configCopy := cm.config
	subscribers := cm.subscribers

	cm.mu.Unlock()

	// 通知所有订阅者
	for _, sub := range subscribers {
		sub(configCopy)
	}

	// 自动保存
	return cm.Save()
}

// UpdateFromJSON 从 JSON 更新配置（前端同步用）
func (cm *ConfigManager) UpdateFromJSON(jsonStr string) error {
	var patch ConfigPatch
	if err := json.Unmarshal([]byte(jsonStr), &patch); err != nil {
		return fmt.Errorf("解析配置 JSON 失败: %w", err)
	}

	return cm.Update(patch)
}

// applyPatch 应用配置补丁
func (cm *ConfigManager) applyPatch(patch ConfigPatch) {
	if patch.APIKey != nil {
		cm.config.APIKey = *patch.APIKey
	}
	if patch.BaseURL != nil {
		cm.config.BaseURL = *patch.BaseURL
	}
	if patch.Model != nil {
		cm.config.Model = *patch.Model
	}
	if patch.Prompt != nil {
		cm.config.Prompt = *patch.Prompt
	}
	if patch.Opacity != nil {
		cm.config.Opacity = *patch.Opacity
	}
	if patch.VoiceListening != nil {
		cm.config.VoiceListening = *patch.VoiceListening
	}
	if patch.ScreenshotMode != nil {
		cm.config.ScreenshotMode = *patch.ScreenshotMode
	}
	if patch.CompressionQuality != nil {
		cm.config.CompressionQuality = *patch.CompressionQuality
	}
	if patch.Sharpening != nil {
		cm.config.Sharpening = *patch.Sharpening
	}
	if patch.Grayscale != nil {
		cm.config.Grayscale = *patch.Grayscale
	}
	if patch.NoCompression != nil {
		cm.config.NoCompression = *patch.NoCompression
	}
	if patch.KeepContext != nil {
		cm.config.KeepContext = *patch.KeepContext
	}
	if patch.InterruptThinking != nil {
		cm.config.InterruptThinking = *patch.InterruptThinking
	}
	if patch.ResumePath != nil {
		cm.config.ResumePath = *patch.ResumePath
	}
	if patch.ResumeContent != nil {
		cm.config.ResumeContent = *patch.ResumeContent
	}
	if patch.UseMarkdownResume != nil {
		cm.config.UseMarkdownResume = *patch.UseMarkdownResume
	}
	if patch.Shortcuts != nil {
		for k, v := range patch.Shortcuts {
			cm.config.Shortcuts[k] = v
		}
	}
}

// Subscribe 订阅配置变更
func (cm *ConfigManager) Subscribe(callback func(Config)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.subscribers = append(cm.subscribers, callback)
}

// ClearResume 清除简历信息
func (cm *ConfigManager) ClearResume() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.config.ResumePath = ""
	cm.config.ResumeBase64 = ""
	cm.config.ResumeContent = ""
}

// SetResumePath 设置简历路径
func (cm *ConfigManager) SetResumePath(path string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.ResumePath = path
}

// SetResumeBase64 设置简历 Base64
func (cm *ConfigManager) SetResumeBase64(base64 string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.ResumeBase64 = base64
}

// SetResumeContent 设置简历内容
func (cm *ConfigManager) SetResumeContent(content string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config.ResumeContent = content
}
