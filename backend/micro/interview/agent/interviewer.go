package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/utils"
)

// InterviewerAgent 面试官 Agent
type InterviewerAgent struct {
	model                   model.ToolCallingChatModel
	interviewerSystemPrompt string
	updateProfilePrompt     string
	scorePrompt             string
}

func NewInterviewerAgent(model model.ToolCallingChatModel) *InterviewerAgent {
	return &InterviewerAgent{
		model:                   model,
		interviewerSystemPrompt: interviewerSystemPrompt,
		updateProfilePrompt:     updateProfilePrompt,
		scorePrompt:             scorePrompt,
	}
}

func (agent *InterviewerAgent) makeAskQuestionMessage(ctx context.Context, state *domain.InterviewState, question *domain.PlannedQuestion, position string) []*schema.Message {
	profileSection := ""
	if state.CandidateProfile != "" {
		profileSection = fmt.Sprintf("\n候选人画像（根据前面的作答动态生成）：\n%s", state.CandidateProfile)
	}
	systemMsg := fmt.Sprintf(agent.interviewerSystemPrompt,
		position,
		state.CurrentQuestionNum,
		state.TotalQuestions,
		state.CurrentDifficulty,
		profileSection,
	)

	// 构建对话历史
	messages := []*schema.Message{
		schema.SystemMessage(systemMsg),
	}

	// 添加之前的问答历史（最近 3 轮）
	historyStart := 0
	if len(state.QAHistory) > 3 {
		historyStart = len(state.QAHistory) - 3
	}
	for _, qa := range state.QAHistory[historyStart:] {
		messages = append(messages,
			schema.AssistantMessage(qa.Question.Content, nil),
			schema.UserMessage(qa.UserAnswer),
		)
	}

	// 添加当前要提出的问题指令
	messages = append(messages,
		schema.UserMessage(fmt.Sprintf("请以面试官的身份直接提出以下面试题，保持简洁，不要加额外的铺垫、背景说明或解释：\n\n%s", question.Content)),
	)

	return messages
}

// AskQuestion 提出面试题目（支持流式输出）
func (agent *InterviewerAgent) AskQuestion(ctx context.Context, state *domain.InterviewState, question *domain.PlannedQuestion, position string) (string, error) {
	// 构造 msg
	msgs := agent.makeAskQuestionMessage(ctx, state, question, position)

	// 调用大模型
	resp, err := agent.model.Generate(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("interviewer: ask question: %w", err)
	}

	// 标注题目来源
	if question.Source != "" && question.Source != "llm" {
		resp.Content += fmt.Sprintf("\n\n`[来源: 题库 %s]`", question.Source)
	} else {
		resp.Content += "\n\n`[来源: LLM 出题]`"
	}

	return resp.Content, nil
}

// AskQuestionStream 提出面试题目（流式输出）
func (agent *InterviewerAgent) AskQuestionStream(ctx context.Context, state *domain.InterviewState, question *domain.PlannedQuestion, position string) (*schema.StreamReader[*schema.Message], error) {
	// 构造 msg
	msgs := agent.makeAskQuestionMessage(ctx, state, question, position)

	// 调用大模型
	stream, err := agent.model.Stream(ctx, msgs)
	if err != nil {
		return nil, fmt.Errorf("interviewer: ask question: %w", err)
	}
	return stream, nil
}

// ScoreAnswer 评估候选人的回答
func (agent *InterviewerAgent) ScoreAnswer(ctx context.Context, question *domain.PlannedQuestion, answer string) (*AnswerScore, error) {
	prompt := fmt.Sprintf(agent.scorePrompt, question.Content, answer, question.Reference)

	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := agent.model.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("interviewer: score answer: %w", err)
	}

	result := &AnswerScore{}
	content := utils.ExtractJSON(resp.Content)
	if err := json.Unmarshal([]byte(content), result); err != nil {
		return nil, fmt.Errorf("interviewer: parse score: %w\nraw: %s", err, resp.Content)
	}

	return result, nil
}

// UpdateCandidateProfile 根据本轮评分结果更新候选人动态画像
func (agent *InterviewerAgent) UpdateCandidateProfile(ctx context.Context, currentProfile string, questionNum int, question *domain.PlannedQuestion, score *AnswerScore) (string, error) {
	prevProfile := "（首次作答，暂无历史画像）"
	if currentProfile != "" {
		prevProfile = "当前画像：\n" + currentProfile
	}

	messages := []*schema.Message{
		schema.UserMessage(fmt.Sprintf(agent.updateProfilePrompt,
			prevProfile,
			questionNum,
			strings.Join(question.Skills, "、"),
			score.Score,
			strings.Join(score.KeyPointsHit, "、"),
			strings.Join(score.KeyPointsMissed, "、"),
		)),
	}

	resp, err := agent.model.Generate(ctx, messages)
	if err != nil {
		return currentProfile, fmt.Errorf("interviewer: update profile: %w", err)
	}

	return resp.Content, nil
}

// FollowUp 基于候选人的实际回答动态生成追问
// question: 当前题目，answer: 候选人的实际回答，feedback: 评分反馈，missedPoints: 遗漏的知识点
func (agent *InterviewerAgent) FollowUp(ctx context.Context, state *domain.InterviewState, question *domain.PlannedQuestion, answer string, feedback string, missedPoints []string, position string) (string, error) {
	profileSection := ""
	if state.CandidateProfile != "" {
		profileSection = fmt.Sprintf("\n候选人画像（根据前面的作答动态生成）：\n%s", state.CandidateProfile)
	}
	systemMsg := fmt.Sprintf(agent.interviewerSystemPrompt,
		position,
		state.CurrentQuestionNum,
		state.TotalQuestions,
		state.CurrentDifficulty,
		profileSection,
	)

	messages := []*schema.Message{
		schema.SystemMessage(systemMsg),
		// 当前这轮的问答（不是历史里的，是正在进行的）
		schema.AssistantMessage(question.Content, nil),
		schema.UserMessage(answer),
	}

	prompt := fmt.Sprintf(`候选人的回答有部分遗漏，请基于以下信息生成一个简短的追问（一句话），引导候选人补充未覆盖的内容。

	评分反馈：%s
	遗漏的知识点：%s
	
	要求：
	- 追问必须基于候选人实际回答的内容来衔接，不要捏造候选人没说过的话
	- 追问要简短自然，像真实面试官一样
	- 不要重复候选人已经回答过的内容`, feedback, strings.Join(missedPoints, "、"))

	messages = append(messages, schema.UserMessage(prompt))

	resp, err := agent.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("interviewer: follow up: %w", err)
	}

	return resp.Content, nil
}

// AnswerScore 回答评分结果
type AnswerScore struct {
	Score           float64  `json:"score"`
	Feedback        string   `json:"feedback"`
	KeyPointsHit    []string `json:"key_points_hit"`
	KeyPointsMissed []string `json:"key_points_missed"`
	ShouldFollowUp  bool     `json:"should_follow_up"`
}

// CollectStreamContent 收集流式输出的完整内容
func CollectStreamContent(stream *schema.StreamReader[*schema.Message]) (string, error) {
	var content string
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return content, err
		}
		content += msg.Content
	}
	return content, nil
}
