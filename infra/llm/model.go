package llm

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/ark"
)

var (
	arkModel *ark.ChatModel
	once     sync.Once
)

func NewArkModel(ctx context.Context, model, apikey string) *ark.ChatModel {
	once.Do(func() {
		var err error
		arkModel, err = ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey: apikey,
			Model:  model,
		})
		if err != nil {
			slog.Error("New Chat Model Failed", "error", err)
			return
		}
	})

	return arkModel
}
