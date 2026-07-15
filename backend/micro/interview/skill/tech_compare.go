/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI Agent知识进阶，后端、AI大模型、场景题面试大全：https://golangstar.cn/
 */

package skill

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
)

// compareTriggers 触发技术对比的关键词
var compareTriggers = []string{
	"对比", "区别", "vs", "VS", "差异",
	"哪个好", "怎么选", "选哪个",
	"和.+比", "跟.+比",
}

// TechCompareSkill 技术对比技能：多维度对比表 → 场景化选型建议
type TechCompareSkill struct {
	chatModel   model.ChatModel
	milvusStore *rag.MilvusStore
	bm25Manager *rag.BM25Manager
}

// NewTechCompareSkill 创建技术对比 Skill
func NewTechCompareSkill(chatModel model.ChatModel, milvusStore *rag.MilvusStore, bm25Manager *rag.BM25Manager) *TechCompareSkill {
	return &TechCompareSkill{
		chatModel:   chatModel,
		milvusStore: milvusStore,
		bm25Manager: bm25Manager,
	}
}

func (t *TechCompareSkill) Name() string { return "tech_compare" }
func (t *TechCompareSkill) Description() string {
	return "技术对比——多维度对比分析、场景化选型建议、面试话术"
}

func (t *TechCompareSkill) Match(input string) bool {
	return ContainsAny(input, compareTriggers)
}

func (t *TechCompareSkill) Handle(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	if state == nil || state.Data == nil || state.Data["phase"] == nil {
		return t.handleStart(ctx, input, state)
	}
	phase, _ := state.Data["phase"].(string)
	switch phase {
	case "advising":
		return t.handleAdvising(ctx, input, state)
	default:
		return t.handleStart(ctx, input, state)
	}
}

// handleStart 启动阶段：生成对比分析
func (t *TechCompareSkill) handleStart(ctx context.Context, input string, existingState *SkillState) (*SkillResponse, error) {
	state := existingState
	if state == nil {
		state = NewSkillState("tech_compare")
	}
	state.Data = make(map[string]interface{})

	// RAG 检索相关参考资料（按用户隔离）
	ragContext := t.fetchRAGContext(ctx, input, state.UserID)

	ragSection := ""
	if ragContext != "" {
		ragSection = fmt.Sprintf("\n\n以下是从题库检索到的相关参考资料，请参考确保对比内容准确：\n%s", ragContext)
	}

	prompt := fmt.Sprintf(`用户想了解：%s

请生成一份结构化的技术对比分析，包含以下三部分：

**第一部分：多维度对比表**
用 Markdown 表格展示，维度包括但不限于：设计理念、性能特点、适用场景、优缺点。

**第二部分：选型建议**
针对 2-3 个常见场景给出具体的选型建议。

**第三部分：面试加分话术**
给一段 50 字左右的面试回答示例，体现对两者的深入理解。

请直接输出 Markdown 格式内容。%s`, input, ragSection)

	resp, err := t.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return nil, fmt.Errorf("tech_compare: 生成对比失败: %w", err)
	}

	state.Data["phase"] = "advising"
	state.Data["analysis"] = resp.Content

	content := fmt.Sprintf("[TechCompare] 正在生成对比分析...\n\n%s\n\n---\n👉 你有具体的使用场景吗？可以给更针对性的建议（输入场景描述，或「结束」退出）",
		resp.Content)

	return &SkillResponse{
		Content:    content,
		Done:       false,
		NextPrompt: "描述你的具体场景获取针对性建议，或输入「结束」退出",
		State:      state,
	}, nil
}

// handleAdvising 追问阶段：根据用户场景给针对性建议
func (t *TechCompareSkill) handleAdvising(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	state.NextRound()

	// 轮次上限
	if state.Round > 5 {
		return t.finish(state)
	}

	// 用户选择结束
	endWords := []string{"结束", "够了", "不用了", "算了", "没有", "不了", "ok", "OK"}
	if ContainsAny(input, endWords) {
		return t.finish(state)
	}

	analysis, _ := state.Data["analysis"].(string)

	prompt := fmt.Sprintf(`基于之前的技术对比分析：
%s

用户描述了自己的具体场景：
%s

请给出针对这个具体场景的选型建议（3-5 句话），说明推荐哪个方案以及原因。`, analysis, input)

	resp, err := t.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return t.finish(state)
	}

	content := fmt.Sprintf("🎯 **针对你的场景：**\n\n%s", resp.Content)

	return &SkillResponse{
		Content: content,
		Done:    true,
		State:   state,
	}, nil
}

// finish 结束
func (t *TechCompareSkill) finish(state *SkillState) (*SkillResponse, error) {
	state.Data["phase"] = "done"
	return &SkillResponse{
		Content: "✅ 技术对比完成！建议将对比表和面试话术整理到笔记中。",
		Done:    true,
		State:   state,
	}, nil
}

// fetchRAGContext 从 RAG 检索相关参考资料（按用户隔离）
func (t *TechCompareSkill) fetchRAGContext(ctx context.Context, topic string, userID int64) string {
	if userID == 0 {
		return ""
	}

	var docs []*schema.Document

	if t.milvusStore != nil {
		if results, err := t.milvusStore.RetrieveByUser(ctx, userID, topic); err == nil {
			docs = append(docs, results...)
		}
	}
	if t.bm25Manager != nil {
		if results, err := t.bm25Manager.Retrieve(ctx, userID, topic); err == nil {
			docs = append(docs, results...)
		}
	}

	if len(docs) == 0 {
		return ""
	}

	var sb strings.Builder
	limit := 3
	if len(docs) < limit {
		limit = len(docs)
	}
	for i := 0; i < limit; i++ {
		sb.WriteString(fmt.Sprintf("---\n%s\n", docs[i].Content))
	}
	return sb.String()
}
