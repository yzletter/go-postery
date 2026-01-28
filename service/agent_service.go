package service

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/document"
	eino_model "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/segmentio/kafka-go"
	ports2 "github.com/yzletter/go-postery/agent/service/ports"
	agentdto "github.com/yzletter/go-postery/dto/agent"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

func (svc *agentService) Chat(ctx context.Context, uid int64, sessionID int64, query string) (agentdto.DTO, error) {
	var empty agentdto.DTO

	// 拉取历史记录
	//messages, err := svc.agentRepo.GetMessagesBySessionID(ctx, sessionID)
	//if err != nil {
	//	if errors.Is(err, repository.ErrServerInternal) {
	//		return empty, errno.ErrServerInternal
	//	}
	//}

	//messages := []string{}

	// RAG 召回
	knowledge, err := svc.agentRepo.Retrieve(ctx, query, 0.5, 3) // 召回分数 > 0.5 的三条
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	//// 聚合信息
	//data := map[string]any{
	//	"document":         knowledge,
	//	"query":            query,
	//	"history_messages": messages,
	//}

	//newQuery, _ := sonic.MarshalString(data)

	// 创建 Agent
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Model:       svc.llmModel,
		Name:        "knowledge_service",
		Description: "知识库助手",
		Instruction: "你是网站的知识库助手，请结合所提供的可靠的论坛文章内容以及历史消息记录以及自己的思考，回答用户的问题",
	})

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
		CheckPointStore: nil,
	})
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	// 进行询问
	iterator := runner.Query(ctx, fmt.Sprintf("资料如下：%s\n用户问题：%s", knowledge, query))

	// 返回结果
	var lastMsg adk.Message
	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			log.Fatal(event.Err)
		}
		msg, err := event.Output.MessageOutput.GetMessage()
		if err != nil {
			log.Fatal(err)
		}
		lastMsg = msg
	}

	if lastMsg == nil {
		return agentdto.DTO{
			SessionID: sessionID,
			Content:   "对不起，这个问题我还在学习中……",
		}, nil
	}

	if lastMsg.Role == schema.Assistant && len(lastMsg.Content) > 0 {
		dto := agentdto.ToDTO(lastMsg, sessionID)
		dto.Documents = knowledge // 参考文献
		return dto, nil
	}

	return agentdto.DTO{
		SessionID: sessionID,
		Content:   "对不起，这个问题我还在学习中……",
	}, nil
}
