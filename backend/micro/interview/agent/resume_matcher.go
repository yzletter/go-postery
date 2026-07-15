package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/utils"
)

// IResumeMatcherAgent 简历与 JD 的匹配度分析 Agent 接口
type IResumeMatcherAgent interface {
	// Match 分析简历与 JD 的匹配度
	Match(ctx context.Context, jd domain.JDAnalysis, resume domain.Resume) (domain.ResumeMatchResult, error)
}

type ResumeMatcherAgent struct {
	model  model.ToolCallingChatModel
	prompt string
}

// NewResumeMatcherAgent 构造函数
func NewResumeMatcherAgent(model model.ToolCallingChatModel) *ResumeMatcherAgent {
	return &ResumeMatcherAgent{
		model:  model,
		prompt: resumeMatcherPrompt,
	}
}

func (agent *ResumeMatcherAgent) Match(ctx context.Context, jdAnalysis domain.JDAnalysis, resume domain.Resume) (domain.ResumeMatchResult, error) {
	jdSummary := formatJDForMatching(jdAnalysis)
	resumeSummary := formatResumeForMatching(resume)

	// 构造 Message
	messages := []*schema.Message{
		schema.SystemMessage(agent.prompt),
		schema.UserMessage(fmt.Sprintf("## 岗位 JD 分析结果\n\n%s\n\n## 候选人简历\n\n%s", jdSummary, resumeSummary)),
	}

	// 调用大模型
	msg, err := agent.model.Generate(ctx, messages)
	if err != nil {
		return domain.ResumeMatchResult{}, err
	}

	// 提取 JSON 进行序列化
	var res domain.ResumeMatchResult

	str := utils.ExtractJSON(msg.Content)
	if err := sonic.UnmarshalString(str, &res); err != nil {
		return domain.ResumeMatchResult{}, err
	}

	// 返回结果
	return res, nil
}

// formatJDForMatching 将 JD 分析结果格式化为便于匹配的文本
func formatJDForMatching(jd domain.JDAnalysis) string {
	data, _ := json.MarshalIndent(jd, "", "  ")
	return string(data)
}

// formatResumeForMatching 将简历格式化为便于匹配的文本
func formatResumeForMatching(resume domain.Resume) string {
	if resume.RawText != "" {
		return resume.RawText
	}
	// 转为便于阅读的缩进形式
	data, _ := json.MarshalIndent(resume, "", "  ")
	return string(data)
}
