/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI Agent知识进阶，后端、AI大模型、场景题面试大全：https://golangstar.cn/
 */

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/utils"
)

// EvaluatorAgent 评估 Agent，负责生成面试评估报告
type EvaluatorAgent struct {
	model  model.ToolCallingChatModel
	prompt string
}

// NewEvaluatorAgent 创建评估 Agent
func NewEvaluatorAgent(model model.ToolCallingChatModel) *EvaluatorAgent {
	return &EvaluatorAgent{
		model:  model,
		prompt: evaluatorPrompt,
	}
}

// Evaluate 生成面试评估报告。userTerminated 表示面试是否由用户主动终止。
func (e *EvaluatorAgent) Evaluate(ctx context.Context, state *domain.InterviewState, position string, candidateName string, userTerminated bool) (*domain.EvaluationReport, error) {
	// 构建面试过程摘要
	var qaText string
	for i, qa := range state.QAHistory {
		qaText += fmt.Sprintf("### 第 %d 题（%s / %s）\n", i+1, qa.Question.Type, qa.Question.Difficulty)
		qaText += fmt.Sprintf("**题目**：%s\n", qa.Question.Content)
		qaText += fmt.Sprintf("**回答**：%s\n", qa.UserAnswer)
		qaText += fmt.Sprintf("**即时得分**：%.0f\n\n", qa.Score)
	}

	terminatedNote := ""
	if userTerminated {
		terminatedNote = fmt.Sprintf("\n\n> **注意：本次面试由候选人主动终止。原计划 %d 道题，实际完成 %d 道题。请在综合评语中说明面试未完成的情况，评估仅基于已作答题目。**\n",
			state.TotalQuestions, len(state.QAHistory))
	}

	userMsg := fmt.Sprintf("## 面试信息\n- 岗位：%s\n- 候选人：%s\n- 计划题目数：%d\n- 实际完成：%d\n- 面试状态：%s%s\n\n## 面试过程\n\n%s",
		position, candidateName, state.TotalQuestions, len(state.QAHistory),
		func() string {
			if userTerminated {
				return "用户主动终止"
			}
			return "正常完成"
		}(), terminatedNote, qaText)

	messages := []*schema.Message{
		schema.SystemMessage(evaluatorPrompt),
		schema.UserMessage(userMsg),
	}

	resp, err := e.model.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("evaluator: generate: %w", err)
	}

	result := &domain.EvaluationReport{}
	content := utils.ExtractJSON(resp.Content)
	if err := json.Unmarshal([]byte(content), result); err != nil {
		return nil, fmt.Errorf("evaluator: parse response: %w\nraw: %s", err, resp.Content)
	}

	result.SessionID = state.SessionID
	result.CandidateName = candidateName
	result.Position = position
	result.CreatedAt = time.Now()

	return result, nil
}

// FormatReport 将评估报告格式化为 Markdown
func FormatReport(report *domain.EvaluationReport) string {
	md := fmt.Sprintf("# 面试评估报告\n\n")
	md += fmt.Sprintf("- **候选人**：%s\n", report.CandidateName)
	md += fmt.Sprintf("- **目标岗位**：%s\n", report.Position)
	md += fmt.Sprintf("- **综合得分**：%.1f / 100（%s）\n", report.OverallScore, report.OverallLevel)
	md += fmt.Sprintf("- **评估时间**：%s\n\n", report.CreatedAt.Format("2006-01-02 15:04"))

	md += "## 各维度得分\n\n"
	md += "| 维度 | 得分 |\n|------|------|\n"
	for dim, score := range report.DimensionScore {
		md += fmt.Sprintf("| %s | %.1f |\n", dim, score)
	}

	md += "\n## 优势\n\n"
	for _, s := range report.Strengths {
		md += fmt.Sprintf("- %s\n", s)
	}

	md += "\n## 待提升\n\n"
	for _, w := range report.Weaknesses {
		md += fmt.Sprintf("- %s\n", w)
	}

	md += "\n## 逐题点评\n\n"
	for i, review := range report.DetailedReview {
		md += fmt.Sprintf("### 第 %d 题（%.0f分）\n", i+1, review.Score)
		md += fmt.Sprintf("**题目**：%s\n\n", review.QuestionContent)
		md += fmt.Sprintf("**点评**：%s\n\n", review.Comment)
	}

	md += fmt.Sprintf("\n## 综合评语\n\n%s\n", report.Summary)

	return md
}
