package node

import (
	"context"
	"errors"
	"log/slog"

	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node/question_scheduler"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
)

var (
	ErrUserQuit           = errors.New("user quit")
	ErrInvalidPhaseChange = errors.New("invalid phase change")
)

type InterviewerNodeBuilder struct {
	LongTermMemory       *memory.LongTermMemory
	InterviewerAgent     *agent.InterviewerAgent
	QuestionPlannerAgent *agent.QuestionPlannerAgent
	Callbacks            FrontendCallbacks
}

func NewInterviewerNodeBuilder(interviewerAgent *agent.InterviewerAgent, questionPlannerAgent *agent.QuestionPlannerAgent, callbacks FrontendCallbacks, longTermMemory *memory.LongTermMemory) *InterviewerNodeBuilder {
	return &InterviewerNodeBuilder{
		LongTermMemory:       longTermMemory,
		InterviewerAgent:     interviewerAgent,
		QuestionPlannerAgent: questionPlannerAgent,
		Callbacks:            callbacks,
	}
}

// BuildInitNode 初始化
func (builder *InterviewerNodeBuilder) BuildInitNode(ctx context.Context, input *RunState) (*RunState, error) {
	// 初始化过
	if input.InterviewState != nil && input.InterviewState.Phase != "" {
		return input, nil
	}

	// 封装 QuestionPlannerAgent 的 AdjustDifficulty 注入
	var adjust = func(cur domain.DifficultyLevel, consequentWrong int, consequentRight int) domain.DifficultyLevel {
		return builder.QuestionPlannerAgent.AdjustDifficulty(domain.InterviewState{
			CurrentDifficulty: cur,
			ConsecutiveRight:  consequentRight,
			ConsecutiveWrong:  consequentWrong,
		})
	}
	questionScheduler := question_scheduler.NewQuestionScheduler(input.QuestionPlan.Questions, adjust)

	// 进行初始化
	input.InterviewState = &domain.InterviewState{
		SessionID:        input.ID,
		AskedQuestionIDs: make(map[int64]struct{}),
		Phase:            domain.PhaseInitDone,
		TotalQuestions:   questionScheduler.Total(),
	}

	// 调度器快照保存
	input.SchedulerSnapshot = questionScheduler.Save()

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	return input, nil
}

// BuildQuestionNode 出题
func (builder *InterviewerNodeBuilder) BuildQuestionNode(ctx context.Context, input *RunState) (*RunState, error) {
	// 校验状态机
	if input.InterviewState.Phase != domain.PhaseInitDone && input.InterviewState.Phase != domain.PhaseUpdateWPDone {
		return input, nil
	}

	// 从快照恢复 Scheduler
	// 封装 QuestionPlannerAgent 的 AdjustDifficulty 注入
	var adjust = func(cur domain.DifficultyLevel, consequentWrong int, consequentRight int) domain.DifficultyLevel {
		return builder.QuestionPlannerAgent.AdjustDifficulty(domain.InterviewState{
			CurrentDifficulty: cur,
			ConsecutiveRight:  consequentRight,
			ConsecutiveWrong:  consequentWrong,
		})
	}
	questionScheduler := question_scheduler.RecoverQuestionScheduler(input.QuestionPlan.Questions, input.SchedulerSnapshot, input.InterviewState.AskedQuestionIDs, adjust)

	// 出题
	var question domain.PlannedQuestion
	var difficulty domain.DifficultyLevel
	var done bool
	for {
		question, difficulty, done = questionScheduler.Next()
		// 题目已经问完
		if done {
			break
		}
		// 判重, 双重保险
		if _, exists := input.InterviewState.AskedQuestionIDs[question.ID]; exists {
			continue
		}
		break
	}

	// 面试结束
	if done {
		// 更新状态机
		input.InterviewState.Phase = domain.PhaseCompleted

		// 更新会话信息
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}

		// 返回
		return input, nil
	}

	// 生成提问语句
	questionText, err := builder.InterviewerAgent.AskQuestion(ctx, input.InterviewState, &question, input.JDAnalysis.Position)
	if err != nil {
		slog.Error("ask interview question failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
		return input, err
	}

	// 更新面试情况
	input.InterviewState.AskedQuestionIDs[question.ID] = struct{}{} // 将当前问题ID放入集合
	input.InterviewState.CurrentQuestion = question
	input.InterviewState.CurrentQuestionNum++ // 先更新计数
	input.InterviewState.CurrentDifficulty = difficulty

	// 发送问题到前端
	builder.Callbacks.OnQuestion(ctx, input.UserID, input.InterviewState.CurrentQuestionNum, questionText)

	// 保存快照
	input.SchedulerSnapshot = questionScheduler.Save()

	// 更新状态机
	input.InterviewState.Phase = domain.PhaseWaitingAnswer

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 返回
	return input, nil
}

// BuildAnswerNode 处理问题回答
func (builder *InterviewerNodeBuilder) BuildAnswerNode(ctx context.Context, input *RunState) (*RunState, error) {
	// 校验状态机
	if input.InterviewState.Phase != domain.PhaseAnswerComing {
		return input, nil
	}

	// 用户主动退出
	if input.UserTerminated {
		input.Status = domain.StatusTerminated

		// 更新状态机
		input.InterviewState.Phase = domain.PhaseUserQuit

		// 更新会话信息
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}

		// 返回
		return input, nil
	}

	// 从快照恢复 Scheduler
	// 封装 QuestionPlannerAgent 的 AdjustDifficulty 注入
	var adjust = func(cur domain.DifficultyLevel, consequentWrong int, consequentRight int) domain.DifficultyLevel {
		return builder.QuestionPlannerAgent.AdjustDifficulty(domain.InterviewState{
			CurrentDifficulty: cur,
			ConsecutiveRight:  consequentRight,
			ConsecutiveWrong:  consequentWrong,
		})
	}
	questionScheduler := question_scheduler.RecoverQuestionScheduler(input.QuestionPlan.Questions, input.SchedulerSnapshot, input.InterviewState.AskedQuestionIDs, adjust)

	question := input.InterviewState.CurrentQuestion
	answer := input.InterviewState.CurrentAnswer

	// 评分
	score, err := builder.InterviewerAgent.ScoreAnswer(ctx, &question, answer)
	if err != nil {
		slog.Error("score answer failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
		return input, err
	}

	// 发送评分到前端
	builder.Callbacks.OnScore(ctx, input.UserID, score)

	// 更新用户画像
	profile, err := builder.InterviewerAgent.UpdateCandidateProfile(ctx, input.InterviewState.CandidateProfile, input.InterviewState.CurrentQuestionNum, &question, score)
	if err != nil {
		// 画像更新失败不影响主流程
		slog.Warn("update candidate profile failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
	}
	input.InterviewState.CandidateProfile = profile

	// 记录问答
	qa := domain.QAPair{
		Question:   question,
		UserAnswer: answer,
		Score:      score.Score,
		Feedback:   score.Feedback,
	}

	// 先将此次 QA 加入历史对话
	input.InterviewState.QAHistory = append(input.InterviewState.QAHistory, qa)

	// 追问
	shouldFollowUp := score.ShouldFollowUp && score.Score >= 30 && score.Score < 80 && len(score.KeyPointsMissed) > 0

	if shouldFollowUp {
		// 进行追问
		followUpText, err := builder.InterviewerAgent.FollowUp(ctx, input.InterviewState, &question, answer, score.Feedback, score.KeyPointsMissed, input.JDAnalysis.Position)
		if err != nil {
			slog.Error("generate follow up question failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
			return input, err
		}

		// 发送追问问题到前端
		builder.Callbacks.OnQuestion(ctx, input.UserID, input.InterviewState.CurrentQuestionNum, "[追问] "+followUpText)

		// 更新状态机
		input.InterviewState.Phase = domain.PhaseWaitingFollowUp

		// 更新会话信息
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}

		// 返回
		return input, nil
	}

	// 动态难度调节（阶段内，由 scheduler 维护；同步到 state 供报告/前端展示）
	questionScheduler.Record(qa.Score)

	// 保存快照
	input.SchedulerSnapshot = questionScheduler.Save()

	// 更新上下文
	input.InterviewState.ConsecutiveRight = questionScheduler.ConsequentRight
	input.InterviewState.ConsecutiveWrong = questionScheduler.ConsequentWrong

	// 更新状态机
	input.InterviewState.Phase = domain.PhaseAnswerDone

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 返回
	return input, nil
}

// BuildFollowUpNode 处理追问回答
func (builder *InterviewerNodeBuilder) BuildFollowUpNode(ctx context.Context, input *RunState) (*RunState, error) {
	// 校验状态机
	if input.InterviewState.Phase != domain.PhaseFollowUpComing {
		return input, nil
	}

	// 用户主动退出
	if input.UserTerminated {
		input.Status = domain.StatusTerminated

		// 更新状态机
		input.InterviewState.Phase = domain.PhaseUserQuit

		// 更新会话信息
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}

		// 返回
		return input, nil
	}

	// 从快照恢复 Scheduler
	// 封装 QuestionPlannerAgent 的 AdjustDifficulty 注入
	var adjust = func(cur domain.DifficultyLevel, consequentWrong int, consequentRight int) domain.DifficultyLevel {
		return builder.QuestionPlannerAgent.AdjustDifficulty(domain.InterviewState{
			CurrentDifficulty: cur,
			ConsecutiveRight:  consequentRight,
			ConsecutiveWrong:  consequentWrong,
		})
	}
	questionScheduler := question_scheduler.RecoverQuestionScheduler(input.QuestionPlan.Questions, input.SchedulerSnapshot, input.InterviewState.AskedQuestionIDs, adjust)

	// 获取主问题回答的 QA 进行修改
	qa := input.InterviewState.QAHistory[len(input.InterviewState.QAHistory)-1]
	// 删除最后一个 QA
	input.InterviewState.QAHistory = input.InterviewState.QAHistory[:len(input.InterviewState.QAHistory)-1]

	answer := input.InterviewState.CurrentAnswer
	question := input.InterviewState.CurrentQuestion

	qa.FollowUpUsed = true
	qa.UserAnswer += "\n[追问回答] " + answer

	// 对追问回答评分并反馈
	followUpScore, fsErr := builder.InterviewerAgent.ScoreAnswer(ctx, &question, answer)
	if fsErr == nil {
		// 发送追问评分到前端
		builder.Callbacks.OnScore(ctx, input.UserID, followUpScore)
	} else {
		slog.Warn("score follow up answer failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", fsErr)
	}

	// 将带 Followup 的 QA 加入历史对话
	input.InterviewState.QAHistory = append(input.InterviewState.QAHistory, qa)

	// 动态难度调节（阶段内，由 scheduler 维护；同步到 state 供报告/前端展示）
	questionScheduler.Record(qa.Score)

	// 保存快照
	input.SchedulerSnapshot = questionScheduler.Save()

	// 更新上下文
	input.InterviewState.ConsecutiveRight = questionScheduler.ConsequentRight
	input.InterviewState.ConsecutiveWrong = questionScheduler.ConsequentWrong

	// 更新状态机
	input.InterviewState.Phase = domain.PhaseFollowUpDone

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 返回
	return input, nil
}

// BuildUpdateWPNode 更新薄弱点
func (builder *InterviewerNodeBuilder) BuildUpdateWPNode(ctx context.Context, input *RunState) (*RunState, error) {
	// 校验状态机
	if input.InterviewState.Phase != domain.PhaseAnswerDone && input.InterviewState.Phase != domain.PhaseFollowUpDone {
		slog.Error("")
		return input, ErrInvalidPhaseChange
	}

	// 获取当轮的主问题
	question := input.InterviewState.CurrentQuestion
	qa := input.InterviewState.QAHistory[len(input.InterviewState.QAHistory)-1]

	// 更新薄弱点
	for _, skill := range question.Skills {
		if err := builder.LongTermMemory.UpdateWeakPoint(ctx, input.UserID, skill, qa.Score); err != nil {
			// 更新失败不影响主进程
			slog.Warn("update weak point failed", "user_id", input.UserID, "session_id", input.ID, "skill", skill, "score", qa.Score, "error", err)
		}
	}

	// 更新状态机
	input.InterviewState.Phase = domain.PhaseUpdateWPDone

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 返回
	return input, nil
}
