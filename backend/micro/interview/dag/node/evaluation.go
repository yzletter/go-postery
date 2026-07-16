package node

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
)

type EvaluationNodeBuilder struct {
	EvaluatorAgent *agent.EvaluatorAgent
	Callbacks      FrontendCallbacks
	LongTermMemory *memory.LongTermMemory
}

func NewEvaluationNodeBuilder(agent *agent.EvaluatorAgent, callbacks FrontendCallbacks, longTermMemory *memory.LongTermMemory) *EvaluationNodeBuilder {
	return &EvaluationNodeBuilder{
		EvaluatorAgent: agent,
		Callbacks:      callbacks,
		LongTermMemory: longTermMemory,
	}
}

func (builder *EvaluationNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
	if input.UserTerminated {
		builder.Callbacks.OnStageChange(ctx, input.UserID, input.ID, StageEvaluationStart, fmt.Sprintf("面试提前终止，正在基于已完成的 %d 道题生成评估报告...", len(input.InterviewState.QAHistory)))
	} else {
		builder.Callbacks.OnStageChange(ctx, input.UserID, input.ID, StageEvaluationStart, "正在生成评估报告...")
		input.Status = domain.StatusEvaluated
	}

	// 评估
	report, err := builder.EvaluatorAgent.Evaluate(ctx, input.InterviewState, input.JDAnalysis.Position, input.Resume.Name, input.UserTerminated)
	if err != nil {
		slog.Error("evaluate interview failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	input.EvaluationReport = report

	reportMD := agent.FormatReport(input.EvaluationReport)
	builder.Callbacks.OnReport(ctx, input.UserID, input.ID, reportMD)
	builder.Callbacks.OnStageChange(ctx, input.UserID, input.ID, StageEvaluationDone, "评估报告生成完成")

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	return input, nil
}
