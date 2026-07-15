package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	eino_utils "github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/mcp"
	"github.com/yzletter/go-postery/backend/utils"
)

// ReviewPlannerAgent 复习规划 Agent —— 全项目唯一真正调用外部工具的 Agent，用 Eino ReAct 实现：
// 模型根据评估报告自主决定是否、用什么关键词调用 GitHub 搜索工具，再综合产出复习计划。
type ReviewPlannerAgent struct {
	model          model.ToolCallingChatModel
	githubSearcher *mcp.GitHubSearcher // 可选，nil 时退化为纯 LLM 生成
}

// NewReviewPlannerAgent 创建复习规划 Agent
func NewReviewPlannerAgent(model model.ToolCallingChatModel) *ReviewPlannerAgent {
	return &ReviewPlannerAgent{model: model}
}

// SetGitHubSearcher 设置 GitHub MCP 搜索器
func (p *ReviewPlannerAgent) SetGitHubSearcher(searcher *mcp.GitHubSearcher) {
	p.githubSearcher = searcher
}

// Plan 根据评估报告生成复习计划
func (p *ReviewPlannerAgent) Plan(ctx context.Context, report *domain.EvaluationReport) (*domain.ReviewPlan, error) {
	reportJSON, _ := json.MarshalIndent(report, "", "  ")
	userMsg := fmt.Sprintf("请根据以下面试评估报告生成复习计划：\n\n%s", string(reportJSON))

	// 优先用 ReAct（带 GitHub 工具，由模型自主调用）；失败或无工具时降级为单轮生成
	content, err := p.generateWithReactAgent(ctx, userMsg)
	if err != nil || strings.TrimSpace(content) == "" {
		if err != nil {
			log.Printf("[ReviewPlanner] ReAct 执行失败，降级为单轮生成: %v", err)
		}
		messages := []*schema.Message{
			schema.SystemMessage(reviewPlannerInstruction),
			schema.UserMessage(userMsg),
		}
		resp, gErr := p.model.Generate(ctx, messages)
		if gErr != nil {
			return nil, fmt.Errorf("review_planner: generate: %w", gErr)
		}
		content = resp.Content
	}

	result := &domain.ReviewPlan{}
	jsonStr := utils.ExtractJSON(content)
	if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
		return nil, fmt.Errorf("review_planner: parse response: %w\nraw: %s", err, content)
	}

	result.SessionID = report.SessionID
	result.CreatedAt = time.Now()

	return result, nil
}

// generateWithReactAgent 用 Eino ReAct（GitHub 工具由模型自主调用）生成复习计划文本。无工具时返回空串以走降级。
func (p *ReviewPlannerAgent) generateWithReactAgent(ctx context.Context, userMsg string) (string, error) {
	if p.githubSearcher == nil {
		return "", nil // 没有 GitHub 工具，直接走降级
	}

	ghTool, err := eino_utils.InferTool(
		"search_github_repos",
		"根据技术关键词搜索 GitHub 上 star 数较多的开源项目与教程，返回项目清单（名称、star 数、链接、简介）。"+
			"为候选人推荐真实可用的学习项目时调用，关键词用英文技术词。",
		p.searchGitHubRepos,
	)
	if err != nil {
		return "", fmt.Errorf("build github tool: %w", err)
	}

	agent, err := react.NewAgent(ctx, &react.AgentConfig{
		ToolCallingModel: p.model,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: []tool.BaseTool{ghTool}},
		MessageModifier: func(_ context.Context, input []*schema.Message) []*schema.Message {
			return append([]*schema.Message{schema.SystemMessage(reviewPlannerInstruction)}, input...)
		},
	})
	if err != nil {
		return "", fmt.Errorf("new react agent: %w", err)
	}

	msg, err := agent.Generate(ctx, []*schema.Message{schema.UserMessage(userMsg)})
	if err != nil {
		return "", err
	}
	return msg.Content, nil
}

// githubSearchReq ReAct 调用 GitHub 工具的入参
type githubSearchReq struct {
	Query string `json:"query" jsonschema:"description=技术关键词，用英文，如 redis distributed-lock"`
}

// searchGitHubRepos 工具实现：按关键词搜索 GitHub 仓库，返回格式化文本
func (p *ReviewPlannerAgent) searchGitHubRepos(ctx context.Context, req githubSearchReq) (string, error) {
	repos, err := p.githubSearcher.SearchRepos(ctx, req.Query+" stars:>100", 5)
	if err != nil || len(repos) == 0 {
		return "未找到相关开源项目。", nil
	}
	var sb strings.Builder
	for i, r := range repos {
		sb.WriteString(fmt.Sprintf("%d. **%s** (%d stars)\n   链接：%s\n   简介：%s\n\n",
			i+1, r.Name, r.Stars, r.URL, r.Desc))
	}
	return sb.String(), nil
}

// FormatReviewPlan 将复习计划格式化为 Markdown
func FormatReviewPlan(plan *domain.ReviewPlan) string {
	md := "# 个性化复习计划\n\n"

	md += "## 薄弱领域\n\n"
	md += "| 领域 | 得分 | 优先级 |\n|------|------|--------|\n"
	for _, area := range plan.WeakAreas {
		md += fmt.Sprintf("| %s | %.1f | %s |\n", area.Topic, area.Score, area.Priority)
	}

	md += "\n## 学习计划\n\n"
	for i, item := range plan.StudyPlan {
		md += fmt.Sprintf("### %d. %s\n\n", i+1, item.Topic)
		md += fmt.Sprintf("**目标**：%s\n\n", item.Objective)
		md += fmt.Sprintf("**预估时间**：%s\n\n", item.TimeEstimate)
		md += "**具体行动**：\n"
		for _, action := range item.Actions {
			md += fmt.Sprintf("- %s\n", action)
		}
		md += "\n"
	}

	if len(plan.Resources) > 0 {
		md += "## 推荐资源\n\n"
		for _, res := range plan.Resources {
			md += fmt.Sprintf("- **[%s](%s)**（%s）：%s\n", res.Title, res.URL, res.Type, res.Desc)
		}
	}

	return md
}
