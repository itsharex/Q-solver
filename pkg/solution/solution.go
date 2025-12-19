package solution

import (
	"Q-Solver/pkg/config"
	"Q-Solver/pkg/llm"
	"Q-Solver/pkg/logger"
	"context"

	"github.com/openai/openai-go"
)

type Callbacks struct {
	EmitEvent func(event string, data ...interface{})
}

type Request struct {
	Config           config.Config
	AudioText        string
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
	if req.Config.APIKey == "" {
		// This case should ideally be handled by the caller or a specific error callback
		// But for now we assume the caller checks this or we return false
		if cb.EmitEvent != nil {
			cb.EmitEvent("require-login")
		}
		return false
	}

	logger.Println("开始解题流程...")

	InitParts := []openai.ChatCompletionContentPartUnionParam{}
	if req.Config.Prompt != "" {
		InitParts = append(InitParts, openai.TextContentPart(req.Config.Prompt))
	}

	// 注入简历
	if req.Config.UseMarkdownResume && req.Config.ResumeContent != "" {
		// 使用 Markdown 简历
		logger.Println("使用 Markdown 简历内容")
		InitParts = append(InitParts, openai.TextContentPart("\n\n# 候选人简历内容：\n"+req.Config.ResumeContent))
	} else if req.ResumeBase64 != "" {
		// 使用 PDF 截图简历
		InitParts = append(InitParts, openai.TextContentPart("\n\n# 候选人简历已作为附件发送，请参考简历内容回答。"))
		InitParts = append(InitParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: "data:application/pdf;base64," + req.ResumeBase64,
		}))
		logger.Println("已注入简历附件 (PDF)")
	}

	//这是准备发送的消息
	var messagesToSend []openai.ChatCompletionMessageParamUnion
	//这是当前轮准备保存的消息
	historyParts := []openai.ChatCompletionContentPartUnionParam{}
	//图片是必定会有的，所以这里加上
	historyParts = append(historyParts, openai.TextContentPart("[用户上传了一张图片，但是为了防止请求体过大，这里用文本替代，**你可以在之前的对话中找到图片内容的描述**。]"))

	//准备用户消息部分
	userParts := []openai.ChatCompletionContentPartUnionParam{}
	userParts = append(userParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
		URL: req.ScreenshotBase64,
	}))

	if len(req.AudioText) > 0 {
		userParts = append(userParts, openai.TextContentPart(req.AudioText))
	}
	currentUserMsg := openai.UserMessage(userParts)
	logger.Println(req.AudioText)

	//先把初始化消息放进去
	if len(InitParts) > 0 {
		messagesToSend = append(messagesToSend, openai.UserMessage(InitParts))
	}

	if req.Config.KeepContext {
		if len(s.chatHistory) > 0 {
			messagesToSend = append(messagesToSend, s.chatHistory...)
		}
	} else {
		s.chatHistory = nil
	}
	messagesToSend = append(messagesToSend, currentUserMsg)

	// 通知前端开始接收流
	if cb.EmitEvent != nil {
		cb.EmitEvent("solution-stream-start")
	}

	answer, err := s.llmProvider.GenerateContentStream(ctx, messagesToSend, func(token string) {
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution-stream-chunk", token)
		}
	})

	if err != nil {
		if ctx.Err() == context.Canceled {
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

	if !req.Config.KeepContext {
		if cb.EmitEvent != nil {
			cb.EmitEvent("solution", answer)
		}
		s.chatHistory = make([]openai.ChatCompletionMessageParamUnion, 0)
		return true
	}

	cleanUserMsg := openai.UserMessage(historyParts)
	// 将瘦身后的消息追加到历史
	s.chatHistory = append(s.chatHistory, cleanUserMsg)
	s.chatHistory = append(s.chatHistory, openai.AssistantMessage(answer))

	if cb.EmitEvent != nil {
		cb.EmitEvent("solution", answer)
	}
	return true
}
