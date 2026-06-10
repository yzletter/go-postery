package llm

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/yzletter/go-postery/backend/conf"
)

var (
	arkModel *ark.ChatModel
	once     sync.Once
)

func NewArkLLMModel(ctx context.Context, config conf.ArkConfig) *ark.ChatModel {
	once.Do(func() {
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
