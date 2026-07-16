package node

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
)

type ResumeMatchNodeBuilder struct {
	ResumeMatcherAgent *agent.ResumeMatcherAgent
	Callbacks          FrontendCallbacks
	LongTermMemory     *memory.LongTermMemory
}

func NewResumeMatchNodeBuilder(agent *agent.ResumeMatcherAgent, callbacks FrontendCallbacks, longTermMemory *memory.LongTermMemory) *ResumeMatchNodeBuilder {
	return &ResumeMatchNodeBuilder{
		ResumeMatcherAgent: agent,
		Callbacks:          callbacks,
		LongTermMemory:     longTermMemory,
	}
}

func (builder *ResumeMatchNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
	// 简历匹配开始
	builder.Callbacks.OnStageChange(ctx, input.UserID, input.ID, StageResumeMatchStart, "正在分析简历匹配度...")

	// 简历匹配
	matchResult, err := builder.ResumeMatcherAgent.Match(ctx, *input.JDAnalysis, *input.Resume)
	if err != nil {
		slog.Error("match resume failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 更新上下文
	input.ResumeMatchResult = &matchResult
	input.Status = domain.StatusResumeMatched

	// 简历匹配结束
	builder.Callbacks.OnStageChange(ctx, input.UserID, input.ID, StageResumeMatchDone, fmt.Sprintf("简历匹配完成，综合匹配度：%.0f%%", matchResult.OverallScore))

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 返回
	return input, nil
}
