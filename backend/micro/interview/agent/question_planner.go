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

type IQuestionPlannerAgent interface {
}

// QuestionPlannerAgent 出题规划 Agent
type QuestionPlannerAgent struct {
	model                  model.ToolCallingChatModel
	directionPlanPrompt    string // 规划出题方向出题方向 prompt
	assembleQuestionPrompt string // 根据出题方向出题 prompt
}

func NewQuestionPlannerAgent(model model.ToolCallingChatModel) *QuestionPlannerAgent {
	return &QuestionPlannerAgent{
		model:                  model,
		directionPlanPrompt:    directionPlannerPrompt,
		assembleQuestionPrompt: questionAssemblerPrompt,
	}
}

// PlanDirections jd + match + weak points -> direction
func (agent *QuestionPlannerAgent) PlanDirections(ctx context.Context, jd domain.JDAnalysis, matchResult domain.ResumeMatchResult, weakPoints string) (domain.QuestionDirectionPlan, error) {
	jdJSON, _ := json.MarshalIndent(jd, "", "  ")
	matchJSON, _ := json.MarshalIndent(matchResult, "", "  ")

	// 构造 msgs
	userMsg := fmt.Sprintf("## JD 分析结果\n\n%s\n\n## 简历匹配结果\n\n%s", string(jdJSON), string(matchJSON))
	if weakPoints != "" {
		userMsg += fmt.Sprintf("\n\n## 候选人历史薄弱点（请针对性加强考察）\n\n%s", weakPoints)
	}

	msgs := []*schema.Message{
		schema.SystemMessage(agent.directionPlanPrompt),
		schema.UserMessage(userMsg),
	}

	// 调用大模型
	resp, err := agent.model.Generate(ctx, msgs)
	if err != nil {
		return domain.QuestionDirectionPlan{}, err
	}

	// 返回结果
	var res domain.QuestionDirectionPlan
	content := utils.ExtractJSON(resp.Content) // 获取内容
	if err := sonic.UnmarshalString(content, &res); err != nil {
		return domain.QuestionDirectionPlan{}, err
	}
	return res, nil
}

/*
"directions": [
	{
		"topic": "考察方向描述（如：Go sync.Map 的并发安全机制）",
		"type": "basic/experience/design",
		"difficulty": "easy/medium/hard",
		"search_query": "题库检索关键词（如：sync.Map 并发）",
		"skills": ["考察的技能点"],
		"context": "简历中相关上下文（experience 类必填，其他类型可为空）"
	}
]
*/

// AssembleQuestion jd + match + direction plan + doc -> question plan
func (agent *QuestionPlannerAgent) AssembleQuestion(ctx context.Context, jd *domain.JDAnalysis, matchResult *domain.ResumeMatchResult, plan domain.QuestionDirectionPlan, docs []string) (domain.QuestionPlan, error) {
	jdJSON, _ := json.MarshalIndent(jd, "", "  ")
	matchJSON, _ := json.MarshalIndent(matchResult, "", "  ")

	var planText string
	for i, direction := range plan.Directions {
		// 基本信息
		planText += fmt.Sprintf("### 方向 %d: %s\n", i+1, direction.Topic)
		planText += fmt.Sprintf("- 类型: %s, 难度: %s, 技能: %v\n", direction.Type, direction.Difficulty, direction.Skills)
		// 简历上下文
		if direction.Context != "" {
			planText += fmt.Sprintf("- 简历上下文: %s\n", direction.Context)
		}

		if i < len(docs) && docs[i] != "" {
			planText += fmt.Sprintf("- 题库匹配原题:\n%s\n", docs[i])
		} else {
			planText += "- 题库匹配: 无匹配，请 LLM 自行出题\n"
		}
		planText += "\n"
	}

	// 构建 msg
	userMsg := fmt.Sprintf("## JD 分析结果\n\n%s\n\n## 简历匹配结果\n\n%s\n\n## 出题方向与题库匹配\n\n%s",
		string(jdJSON), string(matchJSON), planText)

	msg := []*schema.Message{
		schema.SystemMessage(agent.assembleQuestionPrompt),
		schema.UserMessage(userMsg),
	}

	// 调用大模型
	resp, err := agent.model.Generate(ctx, msg)
	if err != nil {
		return domain.QuestionPlan{}, fmt.Errorf("question_planner: assemble questions: %w", err)
	}

	// 返回结果
	var result domain.QuestionPlan
	content := utils.ExtractJSON(resp.Content)
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return domain.QuestionPlan{}, fmt.Errorf("question_planner: parse questions: %w\nraw: %s", err, resp.Content)
	}

	return result, nil
}

func (agent *QuestionPlannerAgent) AdjustDifficulty(state domain.InterviewState) domain.DifficultyLevel {
	// 连续答对 2 题以上 → 提高难度
	if state.ConsecutiveRight >= 2 {
		switch state.CurrentDifficulty {
		case domain.DifficultyEasy:
			return domain.DifficultyMedium
		case domain.DifficultyMedium:
			return domain.DifficultyHard
		default:
			return domain.DifficultyHard
		}
	}

	// 连续答错 2 题以上 → 降低难度
	if state.ConsecutiveWrong >= 2 {
		switch state.CurrentDifficulty {
		case domain.DifficultyHard:
			return domain.DifficultyMedium
		case domain.DifficultyMedium:
			return domain.DifficultyEasy
		default:
			return domain.DifficultyEasy
		}
	}

	// 保持当前难度
	return state.CurrentDifficulty
}
