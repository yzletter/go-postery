/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI Agent知识进阶，后端、AI大模型、场景题面试大全：https://golangstar.cn/
 */

// Package skill 定义 Skill 技能系统的核心接口和数据类型
// Skill 是介于 Tool（MCP 无状态工具）和 Agent（DAG 预编排流程）之间的能力层级：
// 有状态、多轮交互、用户意图动态触发、可插拔
package skill

import (
	"context"
	"strings"
	"time"
)

// Skill 技能接口——每个 Skill 必须实现
type Skill interface {
	// Name 技能名称（唯一标识）
	Name() string
	// Description 技能描述（用于帮助信息展示）
	Description() string
	// Match 判断用户输入是否触发该 Skill
	Match(input string) bool
	// Handle 处理一轮交互
	// input: 用户本轮输入
	// state: 技能状态（首次调用传 nil，后续传上一轮返回的 state）
	Handle(ctx context.Context, input string, state *SkillState) (*SkillResponse, error)
}

// SkillState Skill 的交互状态，跨轮次保持
type SkillState struct {
	SkillName string                 `json:"skill_name"` // 当前激活的 Skill 名称
	UserID    int64                  `json:"user_id"`    // 当前用户 ID（用于 RAG 按用户隔离检索）
	Round     int                    `json:"round"`      // 当前交互轮次（从 1 开始）
	Data      map[string]interface{} `json:"data"`       // Skill 自定义的状态数据
	StartedAt time.Time              `json:"started_at"` // Skill 激活时间
}

// SkillResponse Skill 单轮交互的响应
type SkillResponse struct {
	Content    string      `json:"content"`     // 回复内容（展示给用户）
	Done       bool        `json:"done"`        // 该 Skill 交互是否结束
	NextPrompt string      `json:"next_prompt"` // 引导用户下一步输入的提示语
	State      *SkillState `json:"state"`       // 更新后的状态（传给下一轮 Handle）
}

// NewSkillState 创建一个新的 SkillState
func NewSkillState(skillName string) *SkillState {
	return &SkillState{
		SkillName: skillName,
		Round:     1,
		Data:      make(map[string]interface{}),
		StartedAt: time.Now(),
	}
}

// NextRound 推进到下一轮
func (s *SkillState) NextRound() {
	s.Round++
}

// IsExpired 检查 Skill 会话是否已过期（30 分钟）
func (s *SkillState) IsExpired() bool {
	return time.Since(s.StartedAt) > 30*time.Minute
}

// IsQuitCommand 判断用户输入是否是退出命令
func IsQuitCommand(input string) bool {
	lower := strings.TrimSpace(strings.ToLower(input))
	return lower == "/quit" || lower == "/exit" || lower == "退出" || lower == "结束"
}

// containsAny 检查文本是否包含关键词列表中的任意一个
func ContainsAny(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
