// Package interfaces 定义所有服务的接口抽象
// 用于依赖注入和单元测试
package interfaces

import (
	"Q-Solver/pkg/config"
	"Q-Solver/pkg/screen"
	"Q-Solver/pkg/shortcut"
	"context"

	openai "github.com/openai/openai-go"
)

// ==================== LLM 服务接口 ====================

// LLMProvider 大语言模型提供者接口
type LLMProvider interface {
	// GenerateContentStream 流式生成内容
	GenerateContentStream(ctx context.Context, history []openai.ChatCompletionMessageParamUnion, onToken func(string)) (string, error)
	// GetBalance 获取账户余额
	GetBalance(ctx context.Context) (float64, error)
	// Validate 验证 API Key
	Validate(ctx context.Context) error
	// GetModels 获取可用模型列表
	GetModels(ctx context.Context) ([]string, error)
	// ParseResume 解析简历
	ParseResume(ctx context.Context, resumeBase64 string) (string, error)
	// Transcription 语音转文字
	Transcription(ctx context.Context, audioByte []byte) (string, error)
}

// LLMService LLM 服务接口
type LLMService interface {
	// GetProvider 获取当前 Provider
	GetProvider() LLMProvider
	// UpdateProvider 更新 Provider（配置变更后调用）
	UpdateProvider()
	// GetBalance 获取余额
	GetBalance(ctx context.Context, apiKey string) (float64, error)
	// ValidateAPIKey 验证 API Key
	ValidateAPIKey(ctx context.Context, apiKey string) string
	// GetModels 获取模型列表
	GetModels(ctx context.Context, apiKey string) ([]string, error)
}

// ScreenService 屏幕截图服务接口
type ScreenService interface {
	// Startup 启动服务
	Startup(ctx context.Context)
	// CapturePreview 截图并返回预览
	CapturePreview(quality int, sharpen float64, grayscale bool, noCompression bool, mode string) (screen.PreviewResult, error)
}

// ==================== 简历服务接口 ====================

// ResumeService 简历服务接口
type ResumeService interface {
	// SelectResume 选择简历文件
	SelectResume(ctx context.Context) string
	// ClearResume 清除简历
	ClearResume()
	// GetResumeBase64 获取简历 Base64
	GetResumeBase64() (string, error)
	// ParseResume 解析简历为 Markdown
	ParseResume(ctx context.Context, provider LLMProvider) (string, error)
}

// ==================== 快捷键服务接口 ====================

// ShortcutDelegate 快捷键服务回调接口
type ShortcutDelegate interface {
	TriggerSolve()
	ToggleVisibility()
	ToggleClickThrough()
	MoveWindow(dx, dy int)
	ScrollContent(direction string)
	EmitEvent(eventName string, data ...interface{})
}

// ShortcutService 快捷键服务接口
type ShortcutService interface {
	// Start 启动服务
	Start()
	// Stop 停止服务
	Stop()
	// GetShortcuts 获取当前快捷键配置
	GetShortcuts() map[string]shortcut.KeyBinding
	// SetShortcuts 设置快捷键配置
	SetShortcuts(shortcuts map[string]shortcut.KeyBinding)
	// StartRecording 开始录制快捷键
	StartRecording(action string)
	// StopRecording 停止录制快捷键
	StopRecording()
}

// ==================== 配置服务接口 ====================

// ConfigService 配置服务接口
type ConfigService interface {
	// Load 加载配置
	Load() error
	// Save 保存配置
	Save() error
	// Get 获取当前配置副本
	Get() config.Config
	// GetPtr 获取配置指针
	GetPtr() *config.Config
	// Update 更新配置
	Update(patch config.ConfigPatch) error
	// UpdateFromJSON 从 JSON 更新配置
	UpdateFromJSON(jsonStr string) error
	// Subscribe 订阅配置变更
	Subscribe(callback func(config.Config))
}

// ==================== 状态服务接口 ====================

// StateService 状态管理服务接口
type StateService interface {
	// GetInitStatus 获取初始化状态
	GetInitStatusString() string
	// UpdateInitStatus 更新初始化状态
	UpdateInitStatus(status interface{})
	// IsReady 检查是否就绪
	IsReady() bool
	// ToggleVisibility 切换可见性
	ToggleVisibility() bool
	// ToggleClickThrough 切换鼠标穿透
	ToggleClickThrough() bool
	// MoveWindow 移动窗口
	MoveWindow(dx, dy int)
	// RestoreFocus 恢复焦点
	RestoreFocus()
	// RemoveFocus 移除焦点
	RemoveFocus()
}

// ==================== 任务服务接口 ====================

// TaskService 任务调度服务接口
type TaskService interface {
	// StartTask 开始新任务，返回上下文和任务ID
	StartTask(name string) (context.Context, int64)
	// CompleteTask 标记任务完成
	CompleteTask(taskID int64)
	// CancelCurrentTask 取消当前任务
	CancelCurrentTask() bool
	// IsTaskRunning 检查任务是否在运行
	IsTaskRunning(taskID int64) bool
	// HasRunningTask 检查是否有任务在运行
	HasRunningTask() bool
}

// ==================== 音频服务接口 ====================

// AudioService 音频服务接口
type AudioService interface {
	// Initialize 初始化音频组件
	Initialize(voiceDelegate interface{}) error
	// IsInitialized 检查是否已初始化
	IsInitialized() bool
	// Start 开始语音监听
	Start()
	// Stop 停止语音监听
	Stop()
	// IsRecording 检查是否正在录音
	IsRecording() bool
	// Shutdown 关闭服务
	Shutdown()
}

// ==================== 解题服务接口 ====================

// SolveRequest 解题请求
type SolveRequest struct {
	Config           config.Config
	AudioText        string
	ScreenshotBase64 string
	ResumeBase64     string
}

// SolveCallbacks 解题回调
type SolveCallbacks struct {
	EmitEvent func(event string, data ...interface{})
}

// Solver 解题器接口
type Solver interface {
	// Solve 执行解题
	Solve(ctx context.Context, req SolveRequest, cb SolveCallbacks) bool
	// SetProvider 设置 LLM Provider
	SetProvider(provider LLMProvider)
	// ClearHistory 清空历史
	ClearHistory()
}
