package node

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
)

type JDAnalyzerNodeBuilder struct {
	JDAnalyzerAgent *agent.JDAnalyzerAgent
	Callbacks       FrontendCallbacks
	LongTermMemory  *memory.LongTermMemory
}

func NewJDAnalyzerNodeBuilder(agent *agent.JDAnalyzerAgent, callbacks FrontendCallbacks, longTermMemory *memory.LongTermMemory) *JDAnalyzerNodeBuilder {
	return &JDAnalyzerNodeBuilder{
		JDAnalyzerAgent: agent,
		Callbacks:       callbacks,
		LongTermMemory:  longTermMemory,
	}
}

func (builder *JDAnalyzerNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
	// JD 分析开始
	builder.Callbacks.OnStageChange(ctx, input.UserID, StageJDAnalysisStart, "正在分析岗位 JD...")

	// 进行 JD 分析
	analysis, err := builder.JDAnalyzerAgent.Analyze(ctx, input.JDText)
	if err != nil {
		slog.Error("analyze jd failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 更新上下文
	input.JDAnalysis = &analysis
	input.Status = domain.StatusJDAnalyzed

	// JD 分析结束
	builder.Callbacks.OnStageChange(ctx, input.UserID, StageJDAnalysisDone, fmt.Sprintf("JD 分析完成：%s - %s", analysis.Position, analysis.ExperienceLevel))

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 返回
	return input, nil
}
