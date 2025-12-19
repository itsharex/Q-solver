package audio

import "context"

// VoiceDelegate 定义了 Assistant 需要 App 配合做的事情
type VoiceDelegate interface {
	// UI 更新相关
	EmitEvent(eventName string, data ...interface{})

	// 核心能力相关
	// Solve 让 App 去调用 LLM，返回是否成功
	Solve(ctx context.Context, text string) bool

	// 获取配置 (比如是否允许打断)
	IsInterruptThinkingEnabled() bool
}
