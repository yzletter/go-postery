/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI Agent知识进阶，后端、AI大模型、场景题面试大全：https://golangstar.cn/
 */

package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
)

// quizActions 触发快速测验的动作词
var quizActions = []string{
	"练", "测验", "考考", "出题", "来几道", "做几道", "练几道",
	"快速测验", "出几道", "来道", "做道", "练道",
}

// quizContexts 测验类上下文关键词
var quizContexts = []string{"题", "面试题", "问题"}

// QuickQuizSkill 快速测验技能：RAG 检索题目 → 逐题作答 → 评分 → 总结
type QuickQuizSkill struct {
	chatModel   model.ChatModel
	milvusStore *rag.MilvusStore
	bm25Manager *rag.BM25Manager
}

// NewQuickQuizSkill 创建快速测验 Skill
func NewQuickQuizSkill(chatModel model.ChatModel, milvusStore *rag.MilvusStore, bm25Manager *rag.BM25Manager) *QuickQuizSkill {
	return &QuickQuizSkill{
		chatModel:   chatModel,
		milvusStore: milvusStore,
		bm25Manager: bm25Manager,
	}
}

func (q *QuickQuizSkill) Name() string { return "quick_quiz" }
func (q *QuickQuizSkill) Description() string {
	return "快速测验——输入主题即可练习面试题，即时评分反馈"
}

func (q *QuickQuizSkill) Match(input string) bool {
	lower := strings.ToLower(input)
	if ContainsAny(lower, quizActions) {
		return true
	}
	// "5道Redis题" 这类也能匹配
	if ContainsAny(lower, quizContexts) && hasNumberOrTopic(lower) {
		return true
	}
	return false
}

func (q *QuickQuizSkill) Handle(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	if state == nil || state.Data == nil || state.Data["phase"] == nil {
		return q.handleStart(ctx, input, state)
	}
	phase, _ := state.Data["phase"].(string)
	switch phase {
	case "answering":
		return q.handleAnswer(ctx, input, state)
	default:
		return q.handleStart(ctx, input, state)
	}
}

// handleStart 启动阶段：解析主题和数量，检索题目
func (q *QuickQuizSkill) handleStart(ctx context.Context, input string, existingState *SkillState) (*SkillResponse, error) {
	topic, count := parseQuizInput(input)
	if topic == "" {
		topic = "技术面试"
	}
	if count <= 0 || count > 10 {
		count = 5
	}

	// 复用已有 state（保留 userID），或创建新的
	state := existingState
	if state == nil {
		state = NewSkillState("quick_quiz")
	}
	state.Data = make(map[string]interface{})

	// 从 RAG 检索题目（需要 userID 做用户隔离）
	ragCount := 0
	questions, err := q.retrieveQuestions(ctx, topic, count, state.UserID)
	if err != nil {
		questions = nil
	}
	ragCount = len(questions)

	// RAG 不足的部分用 LLM 生成补充
	if len(questions) < count {
		extra, genErr := q.generateQuestions(ctx, topic, count-len(questions))
		if genErr != nil && len(questions) == 0 {
			return nil, fmt.Errorf("quick_quiz: 获取题目失败: %w", genErr)
		}
		questions = append(questions, extra...)
	}

	state.Data["phase"] = "answering"
	state.Data["topic"] = topic
	state.Data["current"] = 0
	state.Data["scores"] = []interface{}{}

	// 将题目存入 state
	qList := make([]interface{}, len(questions))
	for i, qn := range questions {
		qList[i] = map[string]interface{}{
			"content":    qn.Content,
			"reference":  qn.Reference,
			"difficulty": qn.Difficulty,
		}
	}
	state.Data["questions"] = qList

	// 构建启动提示语（区分题库检索和 LLM 生成）
	var sourceHint string
	if ragCount > 0 && ragCount >= len(questions) {
		sourceHint = fmt.Sprintf("从题库检索到 %d 道题", len(questions))
	} else if ragCount > 0 {
		sourceHint = fmt.Sprintf("从题库检索到 %d 道，LLM 补充生成 %d 道，共 %d 道题", ragCount, len(questions)-ragCount, len(questions))
	} else {
		sourceHint = fmt.Sprintf("题库暂无相关题目，已由 AI 生成 %d 道题", len(questions))
	}

	firstQ := questions[0]
	diffLabel := difficultyLabel(firstQ.Difficulty)
	content := fmt.Sprintf("正在准备「%s」相关面试题...\n%s，快速测验开始！\n\n---\n\n📝 **第 1/%d 题**（%s）\n\n%s",
		topic, sourceHint, len(questions), diffLabel, firstQ.Content)

	return &SkillResponse{
		Content:    content,
		Done:       false,
		NextPrompt: "请输入你的回答（输入 /quit 退出测验）",
		State:      state,
	}, nil
}

// handleAnswer 答题阶段：评分当前回答，输出下一题或总结
func (q *QuickQuizSkill) handleAnswer(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	current, _ := toInt(state.Data["current"])
	questions, _ := state.Data["questions"].([]interface{})
	scores, _ := state.Data["scores"].([]interface{})

	if current >= len(questions) {
		return q.buildSummary(state)
	}

	// 获取当前题目
	qMap, _ := questions[current].(map[string]interface{})
	qContent, _ := qMap["content"].(string)
	qReference, _ := qMap["reference"].(string)

	// 调 LLM 评分
	score, feedback, hit, missed, err := q.scoreAnswer(ctx, qContent, input, qReference)
	if err != nil {
		score = 0
		feedback = "评分失败，已跳过"
		hit = []string{}
		missed = []string{}
	}

	scores = append(scores, score)
	state.Data["scores"] = scores
	state.Data["current"] = current + 1
	state.NextRound()

	// 构建评分反馈
	var sb strings.Builder
	if score >= 70 {
		sb.WriteString(fmt.Sprintf("✅ **得分：%.0f/100**\n\n", score))
	} else {
		sb.WriteString(fmt.Sprintf("⚠️ **得分：%.0f/100**\n\n", score))
	}
	if len(hit) > 0 {
		sb.WriteString("**命中：**" + strings.Join(hit, "、") + "\n\n")
	}
	if len(missed) > 0 {
		sb.WriteString("**遗漏：**" + strings.Join(missed, "、") + "\n\n")
	}
	if feedback != "" {
		sb.WriteString(feedback + "\n")
	}

	// 判断是否还有下一题
	next := current + 1
	if next >= len(questions) {
		// 所有题目答完，输出总结
		sb.WriteString("\n---\n\n")
		summary := q.buildSummaryContent(state, scores)
		sb.WriteString(summary)
		state.Data["phase"] = "done"
		return &SkillResponse{
			Content: sb.String(),
			Done:    true,
			State:   state,
		}, nil
	}

	// 输出下一题（用分隔线区分评分和下一题）
	nextQ, _ := questions[next].(map[string]interface{})
	nextContent, _ := nextQ["content"].(string)
	nextDiff, _ := nextQ["difficulty"].(string)
	sb.WriteString(fmt.Sprintf("\n---\n\n📝 **第 %d/%d 题**（%s）\n\n%s",
		next+1, len(questions), difficultyLabel(nextDiff), nextContent))

	return &SkillResponse{
		Content:    sb.String(),
		Done:       false,
		NextPrompt: "请输入你的回答（输入 /quit 退出测验）",
		State:      state,
	}, nil
}

// buildSummary 构建最终总结
func (q *QuickQuizSkill) buildSummary(state *SkillState) (*SkillResponse, error) {
	scores, _ := state.Data["scores"].([]interface{})
	content := q.buildSummaryContent(state, scores)
	state.Data["phase"] = "done"
	return &SkillResponse{
		Content: content,
		Done:    true,
		State:   state,
	}, nil
}

func (q *QuickQuizSkill) buildSummaryContent(state *SkillState, scores []interface{}) string {
	topic, _ := state.Data["topic"].(string)
	total := len(scores)
	if total == 0 {
		return "📊 快速测验结束，没有作答记录。"
	}

	var sum float64
	scoreStrs := make([]string, total)
	for i, s := range scores {
		v, _ := toFloat64(s)
		sum += v
		scoreStrs[i] = fmt.Sprintf("%.0f", v)
	}
	avg := sum / float64(total)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 **快速测验完成！**\n"))
	sb.WriteString(fmt.Sprintf("主题：%s\n", topic))
	sb.WriteString(fmt.Sprintf("总得分：**%.0f/100**（%d 题平均）\n", avg, total))
	sb.WriteString("各题：" + strings.Join(scoreStrs, " → ") + "\n")

	// 找薄弱题目
	var weakTopics []string
	questions, _ := state.Data["questions"].([]interface{})
	for i, s := range scores {
		v, _ := toFloat64(s)
		if v < 60 && i < len(questions) {
			qMap, _ := questions[i].(map[string]interface{})
			content, _ := qMap["content"].(string)
			if len(content) > 30 {
				content = content[:30] + "..."
			}
			weakTopics = append(weakTopics, content)
		}
	}
	if len(weakTopics) > 0 {
		sb.WriteString("💡 薄弱点：" + strings.Join(weakTopics, "；") + "\n")
		sb.WriteString("建议通过「知识讲解」深入学习相关知识点（输入「讲讲 XXX」即可）。")
	}

	return sb.String()
}

// retrieveQuestions 从 RAG 检索题目（按用户隔离，只检索该用户自己的题库）
func (q *QuickQuizSkill) retrieveQuestions(ctx context.Context, topic string, count int, userID int64) ([]quizQuestion, error) {
	if userID == 0 {
		return nil, nil
	}

	var allDocs []*schema.Document

	// Milvus 向量检索（按用户隔离）
	if q.milvusStore != nil {
		docs, err := q.milvusStore.RetrieveByUser(ctx, userID, topic)
		if err == nil {
			allDocs = append(allDocs, docs...)
		}
	}

	// BM25 关键词检索（按用户隔离）
	if q.bm25Manager != nil {
		docs, err := q.bm25Manager.Retrieve(ctx, userID, topic)
		if err == nil {
			allDocs = append(allDocs, docs...)
		}
	}

	// 去重
	seen := make(map[string]bool)
	var unique []*schema.Document
	for _, d := range allDocs {
		if !seen[d.ID] {
			seen[d.ID] = true
			unique = append(unique, d)
		}
	}

	// 转为 quizQuestion
	var questions []quizQuestion
	for _, d := range unique {
		if len(questions) >= count {
			break
		}
		content, reference := splitContentAndReference(d.Content)
		if content == "" {
			continue
		}
		diff := "medium"
		if d.MetaData != nil {
			if v, ok := d.MetaData["difficulty"].(string); ok && v != "" {
				diff = v
			}
		}
		questions = append(questions, quizQuestion{
			Content:    content,
			Reference:  reference,
			Difficulty: diff,
		})
	}
	return questions, nil
}

// generateQuestions 通过 LLM 生成题目
func (q *QuickQuizSkill) generateQuestions(ctx context.Context, topic string, count int) ([]quizQuestion, error) {
	prompt := fmt.Sprintf(`请生成 %d 道关于「%s」的技术面试题。
要求：
1. 难度从易到难递进
2. 每道题包含题目内容和参考答案要点
3. 只出概念理解、原理分析、场景设计类题目，不要出需要写代码或看代码输出的题目
4. 输出纯 JSON 数组格式：
[
  {"content": "题目内容", "reference": "参考答案要点", "difficulty": "easy/medium/hard"}
]

只输出 JSON，不要其他文字。`, count, topic)

	resp, err := q.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return nil, err
	}

	jsonStr := extractJSON(resp.Content)
	var results []quizQuestion
	if err := json.Unmarshal([]byte(jsonStr), &results); err != nil {
		return nil, fmt.Errorf("parse generated questions: %w", err)
	}
	return results, nil
}

// scoreAnswer 评分（复用面试官 Agent 的评分 Prompt 逻辑）
func (q *QuickQuizSkill) scoreAnswer(ctx context.Context, question, answer, reference string) (float64, string, []string, []string, error) {
	prompt := fmt.Sprintf(`请对候选人的回答进行客观评分和反馈。

题目：%s
候选人回答：%s
参考答案要点：%s

【核心原则】严格基于候选人实际回答的内容进行评分：
- 只认定候选人明确表述出来的知识点
- 候选人没有提到的知识点，一律算作遗漏
- 候选人说"不会"、"不知道"等，得分应为 0-10 分

请输出纯 JSON 格式：
{
  "score": <0-100>,
  "feedback": "简要反馈",
  "key_points_hit": ["命中的知识点"],
  "key_points_missed": ["遗漏的知识点"]
}`, question, answer, reference)

	resp, err := q.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return 0, "", nil, nil, err
	}

	var result struct {
		Score           float64  `json:"score"`
		Feedback        string   `json:"feedback"`
		KeyPointsHit    []string `json:"key_points_hit"`
		KeyPointsMissed []string `json:"key_points_missed"`
	}
	jsonStr := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return 0, "", nil, nil, fmt.Errorf("parse score: %w", err)
	}
	return result.Score, result.Feedback, result.KeyPointsHit, result.KeyPointsMissed, nil
}

// quizQuestion 测验题目
type quizQuestion struct {
	Content    string `json:"content"`
	Reference  string `json:"reference"`
	Difficulty string `json:"difficulty"`
}

// ============================================================
// 辅助函数
// ============================================================

// parseQuizInput 从用户输入中提取主题和题目数量
func parseQuizInput(input string) (string, int) {
	count := 5 // 默认 5 题

	// 提取数字
	re := regexp.MustCompile(`(\d+)\s*道`)
	if matches := re.FindStringSubmatch(input); len(matches) > 1 {
		if n, err := strconv.Atoi(matches[1]); err == nil && n > 0 {
			count = n
		}
	}

	// 移除触发词和数量词，剩下的就是主题
	topic := input
	removeWords := append(quizActions, "道", "面试题", "面试", "相关", "关于", "的", "一下", "帮我", "给我")
	for _, w := range removeWords {
		topic = strings.ReplaceAll(topic, w, "")
	}
	// 移除数字
	topic = regexp.MustCompile(`\d+`).ReplaceAllString(topic, "")
	topic = strings.TrimSpace(topic)

	return topic, count
}

// hasNumberOrTopic 判断输入中是否包含数字或明确的技术主题
func hasNumberOrTopic(input string) bool {
	if regexp.MustCompile(`\d`).MatchString(input) {
		return true
	}
	topics := []string{"go", "redis", "mysql", "kafka", "docker", "k8s",
		"grpc", "http", "并发", "网络", "数据库", "算法", "系统设计", "微服务"}
	return ContainsAny(input, topics)
}

// splitContentAndReference 将 RAG 文档拆分为题目内容和参考答案
func splitContentAndReference(content string) (string, string) {
	if idx := strings.Index(content, "\n参考答案："); idx >= 0 {
		return strings.TrimSpace(content[:idx]), strings.TrimSpace(content[idx+len("\n参考答案："):])
	}
	if idx := strings.Index(content, "参考答案："); idx >= 0 {
		return strings.TrimSpace(content[:idx]), strings.TrimSpace(content[idx+len("参考答案："):])
	}
	return content, ""
}

// difficultyLabel 难度标签
func difficultyLabel(d string) string {
	switch d {
	case "easy":
		return "基础"
	case "medium":
		return "中等"
	case "hard":
		return "困难"
	default:
		return "中等"
	}
}

// extractJSON 从 LLM 输出中提取 JSON
func extractJSON(text string) string {
	// 先尝试 markdown 代码块
	if idx := strings.Index(text, "```json"); idx >= 0 {
		start := idx + 7
		if end := strings.Index(text[start:], "```"); end >= 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	if idx := strings.Index(text, "```"); idx >= 0 {
		start := idx + 3
		if nl := strings.Index(text[start:], "\n"); nl >= 0 {
			start += nl + 1
		}
		if end := strings.Index(text[start:], "```"); end >= 0 {
			return strings.TrimSpace(text[start : start+end])
		}
	}
	// 找第一个 { 或 [ 和最后一个 } 或 ]
	start := strings.IndexAny(text, "{[")
	if start == -1 {
		return text
	}
	endChar := byte('}')
	if text[start] == '[' {
		endChar = ']'
	}
	end := strings.LastIndexByte(text, endChar)
	if end <= start {
		return text
	}
	return text[start : end+1]
}

// toInt 安全转换 interface{} 为 int
func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case float64:
		return int(n), true
	case json.Number:
		i, err := n.Int64()
		return int(i), err == nil
	default:
		return 0, false
	}
}

// toFloat64 安全转换 interface{} 为 float64
func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
