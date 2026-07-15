package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/utils"
)

// IJDAnalyzerAgent JD 分析 Agent 接口
type IJDAnalyzerAgent interface {
	// Analyze 分析文本, 返回结构体结果
	Analyze(ctx context.Context, text string) (domain.JDAnalysis, error)
}

type JDAnalyzerAgent struct {
	model  model.ToolCallingChatModel
	prompt string
}

// NewJDAnalyzerAgent 构造函数
func NewJDAnalyzerAgent(model model.ToolCallingChatModel) *JDAnalyzerAgent {
	return &JDAnalyzerAgent{
		model:  model,
		prompt: jdAnalyzerPrompt,
	}
}

// Analyze 分析文本, 返回结构体结果
func (agent *JDAnalyzerAgent) Analyze(ctx context.Context, text string) (domain.JDAnalysis, error) {
	// 构造 Message
	messages := []*schema.Message{
		schema.SystemMessage(agent.prompt),                       // 系统 Message
		schema.UserMessage(fmt.Sprintf("请分析以下 JD：\n\n%s", text)), // 用户 Message
	}

	// 调用大模型
	msg, err := agent.model.Generate(ctx, messages)
	if err != nil {
		log.Fatal(err)
	}

	// 提取 JSON 进行序列化
	var res domain.JDAnalysis

	str := utils.ExtractJSON(msg.Content)
	if err := sonic.UnmarshalString(str, &res); err != nil {
		return domain.JDAnalysis{}, err
	}
	res.RawJD = text // 原 JD 描述

	// 返回结果
	return res, nil
}
