package node

import (
	"context"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
	"github.com/yzletter/go-postery/backend/micro/interview/model"
	"github.com/yzletter/go-postery/backend/ports"
)

type ReviewPlannerNodeBuilder struct {
	ReviewPlannerAgent *agent.ReviewPlannerAgent
	Callbacks          FrontendCallbacks
	LongTermMemory     *memory.LongTermMemory
	idGen              ports.IDGenerator
}

func NewReviewPlannerNodeBuilder(agent *agent.ReviewPlannerAgent, callbacks FrontendCallbacks, LongTermMemory *memory.LongTermMemory, idGen ports.IDGenerator) *ReviewPlannerNodeBuilder {
	return &ReviewPlannerNodeBuilder{
		ReviewPlannerAgent: agent,
		Callbacks:          callbacks,
		LongTermMemory:     LongTermMemory,
		idGen:              idGen,
	}
}

func (builder *ReviewPlannerNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
	builder.Callbacks.OnStageChange(ctx, input.UserID, StageReviewPlanStart, "正在生成复习计划...")

	// 复习计划
	reviewPlan, err := builder.ReviewPlannerAgent.Plan(ctx, input.EvaluationReport)
	if err != nil {
		slog.Error("plan review failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 更新上下文
	input.ReviewPlan = reviewPlan
	input.Status = domain.StatusCompleted
	input.UpdatedAt = time.Now()

	planMD := agent.FormatReviewPlan(input.ReviewPlan)
	builder.Callbacks.OnReviewPlan(ctx, input.UserID, planMD)
	builder.Callbacks.OnStageChange(ctx, input.UserID, StageReviewPlanDone, "复习计划生成完成")

	reportJSON, _ := sonic.MarshalString(input.EvaluationReport)
	planJSON, _ := sonic.MarshalString(input.ReviewPlan)

	// 面试记录落库
	if err = builder.LongTermMemory.AddInterviewRecord(ctx, input.UserID, model.InterviewRecord{
		ID:             builder.idGen.NextID(),
		UserID:         input.UserID,
		SessionID:      input.ID,
		Position:       input.JDAnalysis.Position,
		OverallScore:   input.EvaluationReport.OverallScore,
		ReportJSON:     reportJSON,
		ReviewPlanJSON: planJSON,
		Date:           time.Now(),
	}); err != nil {
		slog.Warn("add interview record failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
	}

	builder.Callbacks.OnStageChange(ctx, input.UserID, StageCompleted, "面试流程全部完成！")

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	return input, nil
}
