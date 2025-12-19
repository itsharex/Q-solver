package llm

import (
	"Q-Solver/pkg/config"
	"context"
	"fmt"
	"time"
)

type Service struct {
	config   *config.Config
	provider Provider
}

func NewService(cfg *config.Config) *Service {
	s := &Service{
		config: cfg,
	}
	// 初始化时创建 Provider
	s.UpdateProvider()
	return s
}

func (s *Service) UpdateProvider() {
	s.provider = NewOpenAIProvider(s.config.APIKey, s.config.BaseURL, s.config.Model)
}

func (s *Service) GetProvider() Provider {
	return s.provider
}

// GetBalance 获取当前账户余额
func (s *Service) GetBalance(ctx context.Context, apiKey string) (float64, error) {
	// 如果提供了临时的 apiKey，使用临时 provider
	if apiKey != "" && apiKey != s.config.APIKey {
		tempProvider := NewOpenAIProvider(apiKey, s.config.BaseURL, s.config.Model)
		return tempProvider.GetBalance(ctx)
	}

	if s.provider == nil {
		s.UpdateProvider()
	}
	return s.provider.GetBalance(ctx)
}

// ValidateAPIKey 验证 API Key 是否有效
func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) string {
	if apiKey == "" {
		return "API Key 不能为空"
	}

	// 创建一个临时的 provider，只用于验证
	tempProvider := NewOpenAIProvider(apiKey, s.config.BaseURL, s.config.Model)

	// 设置一个超时 context
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := tempProvider.Validate(timeoutCtx)
	if err != nil {
		return fmt.Sprintf("验证失败: %v", err)
	}

	return "" // 成功返回空字符串
}

// GetModels 获取模型列表
func (s *Service) GetModels(ctx context.Context, apiKey string) ([]string, error) {
	var provider *OpenAIProvider

	// 如果提供了临时的 apiKey，使用临时 provider
	if apiKey != "" && apiKey != s.config.APIKey {
		provider = NewOpenAIProvider(apiKey, s.config.BaseURL, s.config.Model)
	} else {
		// 尝试将接口转换为具体类型
		var ok bool
		provider, ok = s.provider.(*OpenAIProvider)
		if !ok {
			return nil, fmt.Errorf("current provider does not support listing models")
		}
	}

	return provider.GetModels(ctx)
}
