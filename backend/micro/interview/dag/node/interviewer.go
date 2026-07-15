package node

//
//import (
//	"context"
//	"errors"
//	"fmt"
//	"log/slog"
//	"time"
//
//	"github.com/yzletter/go-postery/backend/micro/interview/agent"
//	"github.com/yzletter/go-postery/backend/micro/interview/dag/node/question_scheduler"
//	"github.com/yzletter/go-postery/backend/micro/interview/domain"
//	"github.com/yzletter/go-postery/backend/micro/interview/memory"
//)
//
//var (
//	ErrUserQuit = errors.New("user quit")
//)
//
//type InterviewerNodeBuilder struct {
//	LongTermMemory       *memory.LongTermMemory
//	InterviewerAgent     *agent.InterviewerAgent
//	QuestionPlannerAgent *agent.QuestionPlannerAgent
//	Callbacks            FrontendCallbacks
//}
//
//func NewInterviewerNodeBuilder(interviewerAgent *agent.InterviewerAgent, questionPlannerAgent *agent.QuestionPlannerAgent, callbacks FrontendCallbacks, longTermMemory *memory.LongTermMemory) *InterviewerNodeBuilder {
//	return &InterviewerNodeBuilder{
//		LongTermMemory:       longTermMemory,
//		InterviewerAgent:     interviewerAgent,
//		QuestionPlannerAgent: questionPlannerAgent,
//		Callbacks:            callbacks,
//	}
//}
//
//func (builder *InterviewerNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
//	// 创建调度器
//	questionScheduler := question_scheduler.NewQuestionScheduler(input.QuestionPlan.Questions,
//		// 封装 QuestionPlannerAgent 的 AdjustDifficulty 注入
//		func(cur domain.DifficultyLevel, consequentWrong int, consequentRight int) domain.DifficultyLevel {
//			return builder.QuestionPlannerAgent.AdjustDifficulty(domain.InterviewState{
//				CurrentDifficulty: cur,
//				ConsecutiveRight:  consequentRight,
//				ConsecutiveWrong:  consequentWrong,
//			})
//		})
//
//	// 面试开始
//	builder.Callbacks.OnStageChange(ctx, input.UserID, StageInterviewStart, "面试正式开始！")
//
//	// 当前面试的情况
//	input.InterviewState = &domain.InterviewState{
//		SessionID:      input.ID,
//		TotalQuestions: questionScheduler.Total(),
//	}
//
//	// 当前面试的情况
//	userQuit := false
//	terminationNotified := false
//	cnt := 0
//	for {
//		// 抽题
//		question, difficulty, done := questionScheduler.Next()
//		if done {
//			break
//		}
//		cnt++
//		input.InterviewState.CurrentQuestion = cnt          // 当前第几个问题
//		input.InterviewState.CurrentDifficulty = difficulty // 当前难度
//
//		// 生成提问语句
//		questionText, err := builder.InterviewerAgent.AskQuestion(ctx, input.InterviewState, &question, input.JDAnalysis.Position)
//		if err != nil {
//			slog.Error("ask interview question failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
//			return input, err
//		}
//
//		// 发送问题到前端
//		builder.Callbacks.OnQuestion(ctx, input.UserID, cnt, questionText)
//
//		// 答题
//		answer, err := builder.Callbacks.GetUserAnswer(ctx, input.UserID) // 等待前端回答
//		if err != nil {
//			if errors.Is(err, ErrUserQuit) {
//				userQuit = true
//				break
//			}
//			slog.Error("get user answer failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
//			return input, err
//		}
//
//		// 评分
//		score, err := builder.InterviewerAgent.ScoreAnswer(ctx, &question, answer)
//		if err != nil {
//			slog.Error("score answer failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
//			return input, err
//		}
//
//		// 发送评分到前端
//		builder.Callbacks.OnScore(ctx, input.UserID, score)
//
//		// 更新用户画像
//		profile, err := builder.InterviewerAgent.UpdateCandidateProfile(ctx, input.InterviewState.CandidateProfile, cnt, &question, score)
//		if err != nil {
//			// 画像更新失败不影响主流程
//			slog.Warn("update candidate profile failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
//		}
//		input.InterviewState.CandidateProfile = profile
//
//		// 记录问答
//		qa := domain.QAPair{
//			Question:   question,
//			UserAnswer: answer,
//			Score:      score.Score,
//			Feedback:   score.Feedback,
//		}
//
//		// 追问
//		shouldFollowUp := score.ShouldFollowUp && score.Score >= 30 && score.Score < 80 && len(score.KeyPointsMissed) > 0
//		if shouldFollowUp {
//			// 进行追问
//			followUpText, err := builder.InterviewerAgent.FollowUp(ctx, input.InterviewState, &question, answer, score.Feedback, score.KeyPointsMissed, input.JDAnalysis.Position)
//			if err != nil {
//				slog.Error("generate follow up question failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
//				return input, err
//			}
//
//			// 发送追问问题到前端
//			builder.Callbacks.OnQuestion(ctx, input.UserID, cnt, "[追问] "+followUpText)
//
//			// 等待回答
//			answer, err := builder.Callbacks.GetUserAnswer(ctx, input.UserID)
//			if err != nil {
//				if errors.Is(err, ErrUserQuit) {
//					// 追问阶段退出，记录已有的主回答
//					input.InterviewState.QAHistory = append(input.InterviewState.QAHistory, qa)
//					userQuit = true
//					builder.Callbacks.OnStageChange(ctx, input.UserID, StageTerminated, fmt.Sprintf("用户主动终止面试（已完成 %d/%d 题）", len(input.InterviewState.QAHistory), input.InterviewState.TotalQuestions))
//					terminationNotified = true
//					break
//				}
//				slog.Error("get follow up answer failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", err)
//				return input, err
//			}
//
//			qa.FollowUpUsed = true
//			qa.UserAnswer += "\n[追问回答] " + answer
//
//			// 对追问回答评分并反馈
//			followUpScore, fsErr := builder.InterviewerAgent.ScoreAnswer(ctx, &question, answer)
//			if fsErr == nil {
//				// 发送追问评分到前端
//				builder.Callbacks.OnScore(ctx, input.UserID, followUpScore)
//			} else {
//				slog.Warn("score follow up answer failed", "user_id", input.UserID, "session_id", input.ID, "question_id", question.ID, "error", fsErr)
//			}
//		}
//
//		// 将此次 QA 加入历史对话
//		input.InterviewState.QAHistory = append(input.InterviewState.QAHistory, qa)
//
//		// 动态难度调节（阶段内，由 scheduler 维护；同步到 state 供报告/前端展示）
//		questionScheduler.Record(score.Score)
//
//		// 更新上下文
//		input.InterviewState.ConsecutiveRight = questionScheduler.ConsequentRight
//		input.InterviewState.ConsecutiveWrong = questionScheduler.ConsequentWrong
//
//		// 更新薄弱点（长期记忆 → Redis + MySQL）
//		for _, skill := range question.Skills {
//			if err := builder.LongTermMemory.UpdateWeakPoint(ctx, input.UserID, skill, score.Score); err != nil {
//				slog.Warn("update weak point failed", "user_id", input.UserID, "session_id", input.ID, "skill", skill, "score", score.Score, "error", err)
//			}
//		}
//	}
//
//	input.UserTerminated = userQuit
//
//	if userQuit {
//		// 用户主动退出
//		input.Status = domain.StatusTerminated
//		input.UpdatedAt = time.Now()
//		if len(input.InterviewState.QAHistory) == 0 {
//			builder.Callbacks.OnStageChange(ctx, input.UserID, StageTerminated, "面试未作答即终止，不生成评估报告。")
//		} else if !terminationNotified {
//			builder.Callbacks.OnStageChange(ctx, input.UserID, StageTerminated, fmt.Sprintf("用户主动终止面试（已完成 %d/%d 题）", len(input.InterviewState.QAHistory), input.InterviewState.TotalQuestions))
//		}
//	} else {
//		builder.Callbacks.OnStageChange(ctx, input.UserID, StageInterviewDone, fmt.Sprintf("面试答题完成，共完成 %d/%d 题", len(input.InterviewState.QAHistory), input.InterviewState.TotalQuestions))
//	}
//
//	// 更新会话信息
//	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
//		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
//		return input, err
//	}
//
//	return input, nil
//}
