package llm

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	arkModel  *ark.ChatModel
	qwenModel *qwen.ChatModel
	arkOnce   sync.Once
	qwenOnce  sync.Once
)

func NewArkLLMModel(ctx context.Context, config conf.ArkConfig) *ark.ChatModel {
	arkOnce.Do(func() {
		var err error
		arkModel, err = ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: config.APIKey,
			Model:  config.LLMModel,
		})
		if err != nil {
			slog.Error("New Chat Model Failed", "error", err)
			return
		}
	})

	return arkModel
}

// NewQwenLLMModel 千问 LLM 构造函数
func NewQwenLLMModel(ctx context.Context, config conf.QwenConfig) *qwen.ChatModel {
	temperature := float32(0.7)
	qwenOnce.Do(func() {
		var err error
		qwenModel, err = qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			APIKey:      config.APIKey,
			BaseURL:     config.BaseURL,
			Model:       config.LLMModel,
			Temperature: &temperature, // 模型温度
		})
		if err != nil {
			return
		}
	})
	return qwenModel
}
