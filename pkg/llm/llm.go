package llm

import (
	"context"

	openai "github.com/openai/openai-go"
)

// type Message struct {
// 	Role    string
// 	Content string
// 	// 多模态扩展
// 	Images []string // 图片 Base64 列表
// 	Audio  string   // 音频 Base64
// }

// Provider 定义了大模型交互的接口
type Provider interface {
	// GenerateContent 发送提示词和可选的图片到 LLM 并返回响应
	// GenerateContent(ctx context.Context, prompt string, encodedImage string) (string, error)
	// GenerateContentStream 发送流式请求到 LLM
	GenerateContentStream(ctx context.Context, history []openai.ChatCompletionMessageParamUnion, onToken func(string)) (string, error)
	// GetBalance 获取账户余额
	GetBalance(ctx context.Context) (float64, error)
	// Validate 验证 API Key
	Validate(ctx context.Context) error
	// GetModels 获取模型列表
	GetModels(ctx context.Context) ([]string, error)
	// ParseResume 解析简历
	ParseResume(ctx context.Context, resumeBase64 string) (string, error)
	//转录语音
	Transcription(ctx context.Context, audioByte []byte) (string, error)
}
