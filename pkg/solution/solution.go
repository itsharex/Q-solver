package solution

import (
	"Q-Solver/pkg/config"
	"Q-Solver/pkg/llm"
	"Q-Solver/pkg/logger"
	"bytes"
	"context"
	"errors"

	"github.com/openai/openai-go"
)

type Callbacks struct {
	EmitEvent func(event string, data ...interface{})
}

type Request struct {
	Config           config.Config
	ScreenshotBase64 string
	ResumeBase64     string
}

type Solver struct {
	llmProvider llm.Provider
	chatHistory []openai.ChatCompletionMessageParamUnion
}

func NewSolver(provider llm.Provider) *Solver {
	return &Solver{
		llmProvider: provider,
		chatHistory: make([]openai.ChatCompletionMessageParamUnion, 0),
	}
}

func (s *Solver) SetProvider(provider llm.Provider) {
	s.llmProvider = provider
}

func (s *Solver) ClearHistory() {
	s.chatHistory = make([]openai.ChatCompletionMessageParamUnion, 0)
}

func (s *Solver) Solve(ctx context.Context, req Request, cb Callbacks) bool {
	// 1. 检查 API Key
	if req.Config.GetAPIKey() == "" {
		if cb.EmitEvent != nil {
			cb.EmitEvent("require-login")
		}
		return false
	}

	logger.Println("开始解题流程...")

	// 2. 构建 System Prompt
	var systemPrompt bytes.Buffer
	if req.Config.GetPrompt() != "" {
		systemPrompt.WriteString(req.Config.GetPrompt())
	}

	// 如果使用 Markdown 简历，将简历内容追加到 System Prompt
	if req.Config.GetUseMarkdownResume() && req.Config.GetResumeContent() != "" {
		logger.Println("使用 Markdown 简历内容")
		systemPrompt.WriteString("\n\n# 候选人简历内容如下: \n")
		systemPrompt.WriteString(req.Config.GetResumeContent())
	}

	// 3. 构建当前用户消息（包含截图）
	userParts := []openai.ChatCompletionContentPartUnionParam{
		openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: req.ScreenshotBase64,
		}),
	}

	// 如果使用 PDF 简历，将简历附件加入用户消息
	if !req.Config.GetUseMarkdownResume() && req.ResumeBase64 != "" {
		userParts = append(userParts,
			openai.TextContentPart("\n\n# 候选人简历已作为附件发送，请参考简历内容回答。"),
			openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
				URL: "data:application/pdf;base64," + req.ResumeBase64,
			}),
		)
		logger.Println("已注入简历附件 (PDF)")
	}

	currentUserMsg := openai.UserMessage(userParts)

	// 4. 构建最终发送的消息列表
	var messagesToSend []openai.ChatCompletionMessageParamUnion

	if req.Config.GetKeepContext() {
		// 保持上下文模式：使用并更新历史记录
		s.ensureSystemPrompt(systemPrompt.String())
		messagesToSend = append(messagesToSend, s.chatHistory...)
	} else {
		// 不保持上下文模式：每次都是全新对话
		messagesToSend = append(messagesToSend, openai.SystemMessage(systemPrompt.String()))
	}
	messagesToSend = append(messagesToSend, currentUserMsg)

	// 5. 调用 LLM 生成回答
	if cb.EmitEvent != nil {
		cb.EmitEvent("solution-stream-start")
	}

	answer, err := s.llmProvider.GenerateContentStream(ctx, messagesToSend, func(token string) {
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution-stream-chunk", token)
		}
	})

	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			logger.Println("当前任务已中断 (用户产生新输入)")
			if cb.EmitEvent != nil {
				cb.EmitEvent("solution-error", "context canceled")
			}
			return false
		}
		logger.Printf("LLM 请求失败: %v\n", err)
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution-error", err.Error())
		}
		return false
	}

	// 6. 处理结果
	if cb.EmitEvent != nil {
		cb.EmitEvent("solution", answer)
	}

	if req.Config.GetKeepContext() {
		// 保持上下文模式：保存完整的用户消息和助手回复到历史
		s.chatHistory = append(s.chatHistory, currentUserMsg)
		s.chatHistory = append(s.chatHistory, openai.AssistantMessage(answer))
	} else {
		// 不保持上下文模式：清空历史
		s.chatHistory = []openai.ChatCompletionMessageParamUnion{}
	}

	return true
}

// ensureSystemPrompt 确保 chatHistory 的第一条是正确的 System Prompt
func (s *Solver) ensureSystemPrompt(prompt string) {
	if len(s.chatHistory) == 0 {
		s.chatHistory = append(s.chatHistory, openai.SystemMessage(prompt))
		logger.Println("插入 SystemPrompt")
		return
	}

	// 检查第一条是否为系统消息
	if s.chatHistory[0].OfSystem != nil {
		if s.chatHistory[0].OfSystem.Content.OfString.Value != prompt {
			s.chatHistory[0] = openai.SystemMessage(prompt)
			logger.Println("替换 SystemPrompt")
		}
	} else {
		// 第一条不是系统消息，插入到头部
		s.chatHistory = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(prompt)}, s.chatHistory...)
		logger.Println("插入 SystemPrompt 到消息历史头部")
	}
}
