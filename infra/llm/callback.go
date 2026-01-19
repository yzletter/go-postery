package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	cbutils "github.com/cloudwego/eino/utils/callbacks"
)

type LoggerCallbacks struct{}

func (l *LoggerCallbacks) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	slog.Info("[INPUT]", "name", info.Name, "type", info.Type, "component", info.Component, "input", input)
	return ctx
}

func (l *LoggerCallbacks) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	slog.Info("[OUTPUT]", "name", info.Name, "type", info.Type, "component", info.Component, "output", output)
	return ctx
}

func (l *LoggerCallbacks) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	slog.Info("[ERROR]", "name", info.Name, "type", info.Type, "component", info.Component, "error", err)
	return ctx
}

func (l *LoggerCallbacks) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	return ctx
}

func (l *LoggerCallbacks) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return ctx
}

func GetStartCallback() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			slog.Info("[INPUT]", "name", info.Name, "type", info.Type, "component", info.Component, "input", input)
			return ctx
		}).Build()
}

func GetEndCallback() callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			slog.Info("[OUTPUT]", "name", info.Name, "type", info.Type, "component", info.Component, "output", output)
			return ctx
		}).Build()
}

func GetChatModelInputCallback() callbacks.Handler {
	return cbutils.NewHandlerHelper().
		ChatModel(&cbutils.ModelCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
				fmt.Printf("\n[ChatModel Input] component: %s\n", info.Name)
				for i, msg := range input.Messages {
					fmt.Printf("  Message %d [%s]: %s\n", i+1, msg.Role, msg.Content)
					if len(msg.ToolCalls) > 0 {
						fmt.Printf("    Tool Calls: %d\n", len(msg.ToolCalls))
						for j, tc := range msg.ToolCalls {
							fmt.Printf("      %d. %s: %s\n", j+1, tc.Function.Name, tc.Function.Arguments)
						}
					}
				}
				return ctx
			},
		}).Handler()
}

func GetToolInputCallback() callbacks.Handler {
	return cbutils.NewHandlerHelper().
		Tool(&cbutils.ToolCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
				fmt.Printf("\n[Tool Input] component: %s, args: %s\n", info.Name, input.ArgumentsInJSON)
				return ctx
			},
		}).Handler()
}
