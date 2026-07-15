package node

import (
	"context"
	"time"

	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node/question_scheduler"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
)

const (
	JDAnalyzerNodeName   = "JDAnalyzer"
	ResumeMatchNodeName  = "ResumeMatch"
	QuestionPlanNodeName = "QuestionPlan"

	WeakReviewNodeName    = "WeakReview"
	EvaluationNodeName    = "Evaluation"
	ReviewPlannerNodeName = "ReviewPlanner"
)
const (
	InterviewInitNode             = "init"
	InterviewQuestionNode         = "question"
	InterviewAnswerNode           = "answer"
	InterviewFollowUpNode         = "follow_up"
	InterviewUpdateWeakPointsNode = "update_weak_points"
)

// FrontendCallbacks 面试过程回调，用于 CLI/Web 等不同界面
type FrontendCallbacks interface {
	// OnStageChange 面试阶段转换 Callback
	OnStageChange(ctx context.Context, userID int64, stage DAGStage, msg string)
	// OnQuestion 向前端发送问题
	OnQuestion(ctx context.Context, userID int64, questionNum int, content string)
	OnScore(ctx context.Context, userID int64, score *agent.AnswerScore)
	OnReport(ctx context.Context, userID int64, report string)
	OnReviewPlan(ctx context.Context, userID int64, plan string)
	GetUserAnswer(ctx context.Context, userID int64) (string, error)
}

type DAGStage string

const (
	StageJDAnalysisStart  DAGStage = "jd_analysis"
	StageJDAnalysisDone   DAGStage = "jd_analysis_done"
	StageResumeMatchStart DAGStage = "resume_match"
	StageResumeMatchDone  DAGStage = "resume_match_done"
	StageMemoryLoaded     DAGStage = "memory_loaded"
	StageQuestionPlan     DAGStage = "question_plan"
	StageQuestionPlanDone DAGStage = "question_plan_done"
	StageInterviewStart   DAGStage = "interview"
	StageInterviewDone    DAGStage = "interview_done"
	StageWeakReviewStart  DAGStage = "review_weak"
	StageWeakReviewDone   DAGStage = "review_weak_done"
	StageEvaluationStart  DAGStage = "evaluation"
	StageEvaluationDone   DAGStage = "evaluation_done"
	StageReviewPlanStart  DAGStage = "review_plan"
	StageReviewPlanDone   DAGStage = "review_plan_done"
	StageTerminated       DAGStage = "terminated"
	StageCompleted        DAGStage = "completed"
)

type RunState struct {
	UserTerminated    bool                      `json:"user_terminated"`     // 用户主动退出
	ID                int64                     `json:"id"`                  // 面试 ID
	UserID            int64                     `json:"user_id"`             // 用户 ID
	JDText            string                    `json:"jd_text"`             // JD 原始文本
	ResumeText        string                    `json:"resume_text"`         // 简历原始文本
	JDAnalysis        *domain.JDAnalysis        `json:"jd_analysis"`         // JD 分析结构体
	Resume            *domain.Resume            `json:"resume"`              // 简历结构体
	ResumeMatchResult *domain.ResumeMatchResult `json:"resume_match_result"` // 简历匹配结果
	QuestionPlan      *domain.QuestionPlan      `json:"question_plan"`       // 出题计划
	InterviewState    *domain.InterviewState    `json:"interview_state"`     // 面试状态
	EvaluationReport  *domain.EvaluationReport  `json:"evaluation_report"`   // 面试评估
	ReviewPlan        *domain.ReviewPlan        `json:"review_plan"`         // 复盘计划
	Status            domain.SessionStatus      `json:"status"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	SchedulerSnapshot *question_scheduler.SchedulerSnapshot
}
