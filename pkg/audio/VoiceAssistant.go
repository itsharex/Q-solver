package audio

import (
	"Q-Solver/pkg/logger"
	"context"
	"log"
	"strings"
	"sync"
	// "your_project/pkg/logger"
)

type VoiceAssistant struct {
	delegate VoiceDelegate

	mu            sync.Mutex // 替代 reqMu
	isRecording   bool
	thinking      bool
	shouldProcess bool

	// 任务控制
	currentReqID     int64
	cancelCurrentReq context.CancelFunc

	// 数据积累
	fullTranscript strings.Builder
	cancel         context.CancelFunc // 保存取消函数
}

func NewVoiceAssistant(delegate VoiceDelegate) *VoiceAssistant {
	return &VoiceAssistant{
		delegate:      delegate,
		shouldProcess: true, // 默认为 true
	}
}

// Start 启动事件处理循环
func (v *VoiceAssistant) Start(eventChan <-chan AudioEvent) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// 1. 如果之前已经有在跑的任务，先取消掉（防御性编程）
	if v.cancel != nil {
		v.cancel()
	}

	// 2. 创建一个新的上下文
	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel

	go func() {
		// 记得在协程退出时打印日志，方便调试
		logger.Println("VoiceAssistant: 监听协程启动")
		defer logger.Println("VoiceAssistant: 监听协程已退出")

		for {
			select {
			// 情况 A: 收到退出信号 (调用了 v.cancel)
			case <-ctx.Done():
				logger.Println("VoiceAssistant: 监听协程停止")
				return // 直接 return，结束这个协程

			// 情况 B: 收到音频事件
			case event, ok := <-eventChan:
				if !ok {
					// Channel 被关闭了，也应该退出
					return
				}

				// 处理业务逻辑
				switch event.Type {
				case EventStarted:
					v.handleSpeechStart()
				case EventTranscription:
					data, ok := event.Payload.(TranscriptionData)
					if ok {
						v.handleTranscription(data.Text, data.IsFinal)
					}
				case EventStopped:
					v.handleSpeechEnd()
				}
			}
		}
	}()
}

func (v *VoiceAssistant) handleSpeechStart() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.isRecording = true

	interruptEnabled := v.delegate.IsInterruptThinkingEnabled()

	// 场景 A: 正在思考 & 不允许打断
	if v.thinking && !interruptEnabled {
		v.shouldProcess = false
		return
	}

	// 场景 B: 正常处理
	v.shouldProcess = true
	v.delegate.EmitEvent("startRecording", nil)

	// 如果需要打断
	if v.thinking && interruptEnabled {
		// logger.Println("检测到开口，打断 AI...")
		v.interruptExistingTask()
	}
}

func (v *VoiceAssistant) handleTranscription(text string, isFinal bool) {
	text = strings.ReplaceAll(text, " ", "")

	if isFinal {
		v.mu.Lock()
		v.fullTranscript.WriteString(text)
		v.mu.Unlock()
	}

	v.delegate.EmitEvent("asr_result", map[string]any{
		"text":    text,
		"isFinal": isFinal,
	})
}

func (v *VoiceAssistant) handleSpeechEnd() {
	v.mu.Lock()

	v.isRecording = false
	v.delegate.EmitEvent("stopRecording", nil)
	currentText := v.fullTranscript.String()

	// 检查是否忽略
	if !v.shouldProcess {
		v.mu.Unlock()
		return
	}

	if v.thinking && v.cancelCurrentReq != nil {
		if v.delegate.IsInterruptThinkingEnabled() {
			v.interruptExistingTask()
		}
	}

	// 准备新任务
	v.currentReqID++
	reqID := v.currentReqID
	ctx, cancel := context.WithCancel(context.Background())
	v.cancelCurrentReq = cancel
	v.thinking = true

	v.mu.Unlock() // 解锁，开始异步

	// 5. 启动异步任务
	go v.runSolveTask(ctx, reqID, currentText)
}

// 私有辅助方法：执行打断
func (v *VoiceAssistant) interruptExistingTask() {
	if v.cancelCurrentReq != nil {
		v.cancelCurrentReq()
		v.cancelCurrentReq = nil
	}
	v.thinking = false
}

// 私有辅助方法：执行任务协程
func (v *VoiceAssistant) runSolveTask(ctx context.Context, reqID int64, text string) {
	log.Println("runSolveTask", text)
	v.delegate.EmitEvent("start-solving", nil)

	success := v.delegate.Solve(ctx, text)

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.currentReqID != reqID {
		return // 已经被新任务覆盖了
	}

	v.cancelCurrentReq = nil
	v.thinking = false

	if v.isRecording && !v.shouldProcess {
		v.shouldProcess = true
		v.delegate.EmitEvent("startRecording", nil)
	} else {
		v.shouldProcess = true
	}

	if success && ctx.Err() == nil {
		v.fullTranscript.Reset()
	}
}

func (v *VoiceAssistant) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.cancel != nil {
		v.cancel()     // 发送信号给 select -> case <-ctx.Done()
		v.cancel = nil // 清空

	}
}
