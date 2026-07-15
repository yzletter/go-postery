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
)

// projectTriggers 触发项目亮点提炼的关键词
var projectTriggers = []string{
	"项目亮点", "提炼项目", "项目面试", "怎么讲项目",
	"面试话术", "STAR", "star", "项目经历",
	"项目怎么说", "怎么介绍项目", "包装项目",
	"帮我提炼", "帮我包装",
}

// ProjectHighlightSkill 项目亮点提炼技能：STAR 重构 → 亮点提炼 → 模拟追问
type ProjectHighlightSkill struct {
	chatModel model.ChatModel
}

// NewProjectHighlightSkill 创建项目亮点提炼 Skill
func NewProjectHighlightSkill(chatModel model.ChatModel) *ProjectHighlightSkill {
	return &ProjectHighlightSkill{chatModel: chatModel}
}

func (p *ProjectHighlightSkill) Name() string { return "project_highlight" }
func (p *ProjectHighlightSkill) Description() string {
	return "项目亮点包装——STAR 法则重构、技术亮点提炼、模拟面试官追问"
}

func (p *ProjectHighlightSkill) Match(input string) bool {
	return ContainsAny(input, projectTriggers)
}

func (p *ProjectHighlightSkill) Handle(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	if state == nil || state.Data == nil || state.Data["phase"] == nil {
		return p.handleStart(ctx, input, state)
	}
	phase, _ := state.Data["phase"].(string)
	switch phase {
	case "collecting":
		return p.handleCollecting(ctx, input, state)
	case "mock_interview":
		return p.handleMockInterview(ctx, input, state)
	default:
		return p.handleStart(ctx, input, state)
	}
}

// handleStart 启动阶段：判断用户是否已附带项目描述
func (p *ProjectHighlightSkill) handleStart(ctx context.Context, input string, existingState *SkillState) (*SkillResponse, error) {
	state := existingState
	if state == nil {
		state = NewSkillState("project_highlight")
	}
	state.Data = make(map[string]interface{})

	// 判断输入中是否已包含项目描述（超过 50 字认为有实质内容）
	cleaned := input
	for _, t := range projectTriggers {
		cleaned = strings.ReplaceAll(cleaned, t, "")
	}
	cleaned = strings.TrimSpace(cleaned)

	if len([]rune(cleaned)) > 50 {
		// 用户已经贴了项目描述，直接进入分析
		return p.analyzeProject(ctx, cleaned, state)
	}

	// 需要用户输入项目描述
	state.Data["phase"] = "collecting"
	return &SkillResponse{
		Content:    "📋 **项目亮点提炼**\n\n请粘贴你的项目经历描述（包括：项目背景、你的角色、做了什么、用了什么技术、取得了什么成果）。\n\n越详细越好，我会帮你用 STAR 法则重构、提炼面试亮点、预判面试官追问。",
		Done:       false,
		NextPrompt: "请粘贴项目描述（输入 /quit 退出）",
		State:      state,
	}, nil
}

// handleCollecting 收集阶段：收到项目描述后进入分析
func (p *ProjectHighlightSkill) handleCollecting(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	if len([]rune(input)) < 20 {
		return &SkillResponse{
			Content:    "项目描述太短了，请提供更详细的信息（至少包括项目背景和你做了什么）。",
			Done:       false,
			NextPrompt: "请粘贴完整的项目描述（输入 /quit 退出）",
			State:      state,
		}, nil
	}
	return p.analyzeProject(ctx, input, state)
}

// analyzeProject 分析项目并输出结果
func (p *ProjectHighlightSkill) analyzeProject(ctx context.Context, projectDesc string, state *SkillState) (*SkillResponse, error) {
	state.Data["project_desc"] = projectDesc
	state.Data["phase"] = "mock_interview"

	prompt := fmt.Sprintf(`你是一位资深技术面试辅导专家。请根据以下项目描述，输出三部分内容：

项目描述：
%s

---

**第一部分：STAR 结构重组**
按 Situation（背景）→ Task（任务）→ Action（行动）→ Result（成果）重新组织项目描述，让表述更清晰有力。

**第二部分：技术亮点提炼（3-5 个）**
从项目中提取可以在面试中展开讲的技术亮点，每个亮点给出：
- 一句话亮点（面试中的开场白）
- 30 秒展开话术（如何深入讲解）

**第三部分：预判面试官追问（3 个）**
预测面试官最可能针对这个项目追问的 3 个问题，每个给出回答思路提示。

请直接输出内容，使用 Markdown 格式。`, projectDesc)

	resp, err := p.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return nil, fmt.Errorf("project_highlight: 分析失败: %w", err)
	}

	state.Data["analysis"] = resp.Content
	state.Data["mock_current"] = 0

	content := fmt.Sprintf("[ProjectHighlight] 正在分析项目亮点...\n\n%s\n\n---\n👉 要不要我扮演面试官，模拟追问一下？（输入「好的」开始，或「跳过」结束）",
		resp.Content)

	return &SkillResponse{
		Content:    content,
		Done:       false,
		NextPrompt: "输入「好的」开始模拟追问，或「跳过」结束（输入 /quit 退出）",
		State:      state,
	}, nil
}

// handleMockInterview 模拟追问阶段
func (p *ProjectHighlightSkill) handleMockInterview(ctx context.Context, input string, state *SkillState) (*SkillResponse, error) {
	mockCurrent, _ := toInt(state.Data["mock_current"])
	state.NextRound()

	// 轮次上限
	if state.Round > 15 {
		return p.finishMock(state)
	}

	// 用户选择跳过
	skipWords := []string{"跳过", "结束", "不用了", "算了", "不了"}
	if ContainsAny(input, skipWords) {
		return p.finishMock(state)
	}

	projectDesc, _ := state.Data["project_desc"].(string)

	if mockCurrent == 0 && ContainsAny(input, []string{"好的", "好", "开始", "可以", "来吧", "ok", "OK"}) {
		// 开始第一个追问
		question, err := p.generateMockQuestion(ctx, projectDesc, mockCurrent, "")
		if err != nil {
			return p.finishMock(state)
		}
		state.Data["mock_current"] = 1
		state.Data["last_question"] = question
		return &SkillResponse{
			Content:    fmt.Sprintf("🎤 **模拟追问 1/3**\n\n%s", question),
			Done:       false,
			NextPrompt: "请回答（输入「跳过」结束模拟追问）",
			State:      state,
		}, nil
	}

	// 用户回答了追问，给反馈并出下一个追问
	lastQuestion, _ := state.Data["last_question"].(string)
	feedback, err := p.evaluateMockAnswer(ctx, projectDesc, lastQuestion, input)
	if err != nil {
		feedback = "回答不错，继续保持。"
	}

	if mockCurrent >= 3 {
		// 追问完毕
		content := fmt.Sprintf("💬 **反馈：**\n%s\n\n---\n✅ 模拟追问完成！建议多练习几次，让回答更自然流畅。", feedback)
		return &SkillResponse{
			Content: content,
			Done:    true,
			State:   state,
		}, nil
	}

	// 生成下一个追问
	nextQuestion, err := p.generateMockQuestion(ctx, projectDesc, mockCurrent, input)
	if err != nil {
		return p.finishMock(state)
	}
	state.Data["mock_current"] = mockCurrent + 1
	state.Data["last_question"] = nextQuestion

	content := fmt.Sprintf("💬 **反馈：**\n%s\n\n🎤 **模拟追问 %d/3**\n\n%s",
		feedback, mockCurrent+1, nextQuestion)

	return &SkillResponse{
		Content:    content,
		Done:       false,
		NextPrompt: "请回答（输入「跳过」结束模拟追问）",
		State:      state,
	}, nil
}

// finishMock 结束模拟追问
func (p *ProjectHighlightSkill) finishMock(state *SkillState) (*SkillResponse, error) {
	state.Data["phase"] = "done"
	return &SkillResponse{
		Content: "✅ 项目亮点提炼完成！建议将 STAR 结构和技术亮点整理到面试准备笔记中。",
		Done:    true,
		State:   state,
	}, nil
}

// generateMockQuestion 生成模拟追问
func (p *ProjectHighlightSkill) generateMockQuestion(ctx context.Context, projectDesc string, idx int, lastAnswer string) (string, error) {
	contextInfo := ""
	if lastAnswer != "" {
		contextInfo = fmt.Sprintf("\n候选人上一轮的回答：%s", lastAnswer)
	}

	prompt := fmt.Sprintf(`你是一位资深技术面试官，正在针对候选人的项目经历进行追问。

项目描述：
%s
%s

这是第 %d 个追问。请基于项目内容生成一个深入的追问问题，考察候选人对技术细节的理解。
要求：一句话，简短直接，像真实面试官一样。`, projectDesc, contextInfo, idx+1)

	resp, err := p.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// evaluateMockAnswer 评估模拟追问的回答
func (p *ProjectHighlightSkill) evaluateMockAnswer(ctx context.Context, projectDesc, question, answer string) (string, error) {
	prompt := fmt.Sprintf(`作为技术面试辅导专家，请对候选人的回答给出简短反馈（2-3句话）。

项目背景：%s
面试官问题：%s
候选人回答：%s

反馈要求：
- 指出回答好的地方
- 指出可以改进的地方
- 给一个具体的改进建议`, projectDesc, question, answer)

	resp, err := p.chatModel.Generate(ctx, []*schema.Message{schema.UserMessage(prompt)})
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
