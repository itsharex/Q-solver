package audio

import (
	"Q-Solver/pkg/logger"
	"sync"
)

// AudioDelegate 音频管理器需要的回调接口
type AudioDelegate interface {
	// EmitEvent 发送事件到前端
	EmitEvent(eventName string, data ...interface{})
	// OnInitStatusChange 初始化状态变更回调
	OnAudioInitStatusChange(status string)
}

// Manager 音频管理器，封装所有音频相关功能
type Manager struct {
	delegate AudioDelegate

	// 核心组件
	recorder     *AudioRecorder
	voiceService *VoiceAssistant
	eventChan    chan AudioEvent

	// 状态管理
	mu             sync.Mutex
	initOnce       sync.Once
	initialized    bool
	isRecording    bool
	voiceAvailable bool // 语音功能是否可用
}

// NewManager 创建音频管理器
func NewManager(delegate AudioDelegate) *Manager {
	return &Manager{
		delegate:  delegate,
		eventChan: make(chan AudioEvent, 100),
	}
}

// Initialize 初始化音频组件（只会执行一次）
func (m *Manager) Initialize(voiceDelegate VoiceDelegate) error {
	var initErr error

	m.initOnce.Do(func() {
		if m.delegate != nil {
			m.delegate.OnAudioInitStatusChange("loading-model")
		}

		recorder, err := NewAudioRecorder(m.eventChan)
		if err != nil {
			// 模型不存在，禁用语音功能但不阻止程序启动
			logger.Printf("语音功能不可用: %v，将禁用语音监听功能", err)
			m.mu.Lock()
			m.voiceAvailable = false
			m.initialized = true // 标记为已初始化，只是语音不可用
			m.mu.Unlock()
			if m.delegate != nil {
				m.delegate.OnAudioInitStatusChange("ready")
			}
			return
		}

		m.recorder = recorder
		m.voiceService = NewVoiceAssistant(voiceDelegate)

		m.mu.Lock()
		m.voiceAvailable = true
		m.initialized = true
		m.mu.Unlock()

		if m.delegate != nil {
			m.delegate.OnAudioInitStatusChange("ready")
		}
		logger.Println("音频组件初始化完成")
	})

	return initErr
}

// IsInitialized 检查是否已初始化
func (m *Manager) IsInitialized() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.initialized
}

// Start 开始语音监听
func (m *Manager) Start() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRecording {
		logger.Println("语音监听已在运行中，跳过重复启动请求")
		return
	}

	if !m.initialized {
		logger.Println("音频组件未初始化，无法启动")
		return
	}

	if !m.voiceAvailable {
		logger.Println("语音功能不可用（模型文件缺失），跳过启动")
		return
	}

	// 启动语音服务
	if m.voiceService != nil {
		m.voiceService.Start(m.eventChan)
	}

	// 启动录音器
	if m.recorder != nil {
		if err := m.recorder.Start(); err != nil {
			logger.Printf("启动录音器失败: %v\n", err)
			return
		}
		logger.Println("【状态变更】语音监听已启动")
	}

	m.isRecording = true
}

// Stop 停止语音监听
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRecording {
		return
	}

	if m.recorder != nil {
		m.recorder.Stop()
		logger.Println("【状态变更】语音监听已停止")
	}

	if m.voiceService != nil {
		m.voiceService.Stop()
	}

	m.isRecording = false
}

// IsRecording 检查是否正在录音
func (m *Manager) IsRecording() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isRecording
}

// IsVoiceAvailable 检查语音功能是否可用
func (m *Manager) IsVoiceAvailable() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.voiceAvailable
}

// GetEventChannel 获取事件通道（用于外部订阅）
func (m *Manager) GetEventChannel() <-chan AudioEvent {
	return m.eventChan
}

// Shutdown 关闭音频管理器
func (m *Manager) Shutdown() {
	m.Stop()
	// 关闭事件通道
	close(m.eventChan)
}
