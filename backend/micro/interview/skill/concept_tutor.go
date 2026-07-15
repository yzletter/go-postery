/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI Agent知识进阶，后端、AI大模型、场景题面试大全：https://golangstar.cn/
 */

package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
)

// tutorTriggers 触发知识讲解的关键词
var tutorTriggers = []string{
	"讲讲", "讲一下", "讲解", "解释一下", "给我讲",
	"帮我理解", "什么是", "原理", "底层",
	"怎么实现的", "如何实现", "机制是什么",
}

// ConceptTutorSkill 知识点讲解技能：分步讲解 + 互动确认
type ConceptTutorSkill struct {
	chatModel   model.ChatModel
	milvusStore *rag.MilvusStore
	bm25Manager *rag.BM25Manager
}

// NewConceptTutorSkill 创建知识讲解 Skill
func NewConceptTutorSkill(chatModel model.ChatModel, milvusStore *rag.MilvusStore, bm25Manager *rag.BM25Manager) *ConceptTutorSkill {
	return &ConceptTutorSkill{
		chatModel:   chatModel,
		milvusStore: milvusStore,
		bm25Manager: bm25Manager,
	}
}

func (c *ConceptTutorSkill) Name() string { return "concept_tutor" }
func (c *ConceptTutorSkill) Description() string {
	return "知识点讲解——分步教学、互动确认，帮你深入理解技术概念"
}

func (c *ConceptTutorSkill) Match(input string) bool {
	lower := strings.ToLower(input)
	// 必须包含教学类关键词，且不包含测验类动作词（避免和 QuickQuiz 冲突）
	return ContainsAny(lower, tutorTriggers) && !ContainsAny(lower, quizActions)
}

func (c *ConceptTutorSkill) Handle(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	if state == nil || state.Data == nil || state.Data["phase"] == nil {
		return c.handleStart(ctx, input, state)
	}
	phase, _ := state.Data["phase"].(string)
	switch phase {
	case "teaching":
		return c.handleTeaching(ctx, input, state)
	default:
		return c.handleStart(ctx, input, state)
	}
}

// handleStart 启动阶段：提取主题 → RAG 检索参考 → LLM 拆解要点 → 输出第一个要点
func (c *ConceptTutorSkill) handleStart(ctx context.Context, input string, existingState *SkillState) (*SkillResponse, error) {
	topic := parseTutorTopic(input)
	if topic == "" {
		topic = input
	}

	// 复用已有 state（保留 userID），或创建新的
	state := existingState
	if state == nil {
		state = NewSkillState("concept_tutor")
	}
	state.Data = make(map[string]interface{})

	// RAG 检索相关参考资料（按用户隔离）
	ragContext := c.fetchRAGContext(ctx, topic, state.UserID)

	// LLM 拆解知识点为核心要点
	keyPoints, err := c.decomposeKeyPoints(ctx, topic, ragContext)
	if err != nil {
		return nil, fmt.Errorf("concept_tutor: 拆解知识点失败: %w", err)
	}
	if len(keyPoints) == 0 {
		keyPoints = []string{"核心概念与定义", "工作原理", "应用场景与最佳实践"}
	}

	state.Data["phase"] = "teaching"
	state.Data["topic"] = topic
	state.Data["key_points"] = toInterfaceSlice(keyPoints)
	state.Data["current"] = 0
	state.Data["retry_count"] = 0
	state.Data["rag_context"] = ragContext

	// 生成第一个要点的讲解
	explanation, err := c.explainPoint(ctx, topic, keyPoints[0], ragContext, false)
	if err != nil {
		return nil, fmt.Errorf("concept_tutor: 生成讲解失败: %w", err)
	}

	content := fmt.Sprintf("[ConceptTutor] 正在准备「%s」的讲解内容...\n\n📖 「%s」共 %d 个核心要点，我们逐个来看。\n\n🔹 **要点 1/%d：%s**\n%s",
		topic, topic, len(keyPoints), len(keyPoints), keyPoints[0], explanation)

	return &SkillResponse{
		Content:    content,
		Done:       false,
		NextPrompt: "理解了请输入「继续」，不清楚可以说「再讲讲」（输入 /quit 退出）",
		State:      state,
	}, nil
}

// handleTeaching 教学阶段：根据用户反馈继续下一个要点或重讲
func (c *ConceptTutorSkill) handleTeaching(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	topic, _ := state.Data["topic"].(string)
	keyPointsRaw, _ := state.Data["key_points"].([]interface{})
	current, _ := toInt(state.Data["current"])
	retryCount, _ := toInt(state.Data["retry_count"])
	ragContext, _ := state.Data["rag_context"].(string)

	keyPoints := toStringSlice(keyPointsRaw)
	if current >= len(keyPoints) {
		return c.buildConclusion(state, topic, keyPoints)
	}

	state.NextRound()

	// 轮次上限检查
	if state.Round > 20 {
		return c.buildConclusion(state, topic, keyPoints)
	}

	// 判断用户是否理解了
	understood := isUnderstandResponse(input)

	if understood || retryCount >= 2 {
		// 进入下一个要点
		next := current + 1
		state.Data["current"] = next
		state.Data["retry_count"] = 0

		if next >= len(keyPoints) {
			return c.buildConclusion(state, topic, keyPoints)
		}

		explanation, err := c.explainPoint(ctx, topic, keyPoints[next], ragContext, false)
		if err != nil {
			explanation = "（讲解生成失败，请稍后重试）"
		}

		content := fmt.Sprintf("🔹 **要点 %d/%d：%s**\n%s",
			next+1, len(keyPoints), keyPoints[next], explanation)

		return &SkillResponse{
			Content:    content,
			Done:       false,
			NextPrompt: "理解了请输入「继续」，不清楚可以说「再讲讲」（输入 /quit 退出）",
			State:      state,
		}, nil
	}

	// 用户不理解，换个方式重讲
	state.Data["retry_count"] = retryCount + 1
	explanation, err := c.explainPoint(ctx, topic, keyPoints[current], ragContext, true)
	if err != nil {
		explanation = "（讲解生成失败，请稍后重试）"
	}

	content := fmt.Sprintf("🔄 换个方式来讲「%s」：\n%s", keyPoints[current], explanation)

	return &SkillResponse{
		Content:    content,
		Done:       false,
		NextPrompt: "这次理解了吗？（输入「继续」进入下一个要点，或继续提问）",
		State:      state,
	}, nil
}

// buildConclusion 收尾阶段：输出总结
func (c *ConceptTutorSkill) buildConclusion(state *SkillState, topic string, keyPoints []string) (*SkillResponse, error) {
	state.Data["phase"] = "done"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n✅ 「%s」讲解完成！\n\n", topic))
	sb.WriteString("📋 **要点回顾：**\n")
	for i, kp := range keyPoints {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, kp))
	}
	sb.WriteString(fmt.Sprintf("\n💡 建议通过快速测验检验学习效果（输入「来几道 %s 的题」即可）。", topic))

	return &SkillResponse{
		Content: sb.String(),
		Done:    true,
		State:   state,
	}, nil
}

// decomposeKeyPoints 用 LLM 将知识点拆解为核心要点
func (c *ConceptTutorSkill) decomposeKeyPoints(ctx context.Context, topic, ragContext string) ([]string, error) {
	ragSection := ""
	if ragContext != "" {
		ragSection = fmt.Sprintf("\n\n以下是从题库检索到的相关参考资料，请参考：\n%s", ragContext)
	}

	prompt := fmt.Sprintf(`请将「%s」这个技术知识点拆解为 3-5 个核心要点，用于分步教学。
要求：
1. 每个要点是一个简短的标题（10字以内）
2. 从基础到深入排列
3. 覆盖：定义/原理 → 实现机制 → 应用场景/面试考点
%s

输出纯 JSON 数组，如：["三色标记法", "写屏障机制", "GC触发条件", "调优实践"]`, topic, ragSection)

	resp, err := c.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(resp.Content)
	var points []string
	if err := json.Unmarshal([]byte(jsonStr), &points); err != nil {
		return nil, fmt.Errorf("parse key points: %w", err)
	}
	return points, nil
}

// explainPoint 生成某个要点的讲解
func (c *ConceptTutorSkill) explainPoint(ctx context.Context, topic, point, ragContext string, useAnalogy bool) (string, error) {
	style := "清晰直接地讲解，可以用简短代码示例辅助说明"
	if useAnalogy {
		style = "用生活中的类比或具体的例子来解释，避免抽象术语"
	}

	ragSection := ""
	if ragContext != "" {
		ragSection = fmt.Sprintf("\n\n参考资料：\n%s", ragContext)
	}

	prompt := fmt.Sprintf(`请讲解「%s」中「%s」这个要点。

要求：
- %s
- 控制在 200 字以内
- 面向技术面试准备，突出面试中的高频考点
- 不要输出标题，直接讲内容%s`, topic, point, style, ragSection)

	resp, err := c.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// fetchRAGContext 从 RAG 检索相关参考资料（按用户隔离）
func (c *ConceptTutorSkill) fetchRAGContext(ctx context.Context, topic string, userID int64) string {
	if userID == 0 {
		return ""
	}

	var docs []*schema.Document

	if c.milvusStore != nil {
		if results, err := c.milvusStore.RetrieveByUser(ctx, userID, topic); err == nil {
			docs = append(docs, results...)
		}
	}
	if c.bm25Manager != nil {
		if results, err := c.bm25Manager.Retrieve(ctx, userID, topic); err == nil {
			docs = append(docs, results...)
		}
	}

	if len(docs) == 0 {
		return ""
	}

	// 取前 3 条作为参考
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

// ============================================================
// 辅助函数
// ============================================================

// parseTutorTopic 从用户输入中提取知识点主题
func parseTutorTopic(input string) string {
	topic := input
	removeWords := append(tutorTriggers, "帮我", "给我", "一下", "的", "吗", "呢", "？", "?")
	for _, w := range removeWords {
		topic = strings.ReplaceAll(topic, w, "")
	}
	return strings.TrimSpace(topic)
}

// isUnderstandResponse 判断用户回复是否表示理解了
func isUnderstandResponse(input string) bool {
	positiveWords := []string{
		"继续", "下一个", "明白", "理解", "懂了", "知道了",
		"ok", "OK", "好的", "可以", "嗯", "对", "是的", "next",
	}
	return ContainsAny(input, positiveWords)
}

// toInterfaceSlice 将 string 切片转为 interface{} 切片（存入 SkillState.Data）
func toInterfaceSlice(ss []string) []interface{} {
	result := make([]interface{}, len(ss))
	for i, s := range ss {
		result[i] = s
	}
	return result
}

// toStringSlice 将 interface{} 切片转回 string 切片
func toStringSlice(raw []interface{}) []string {
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
