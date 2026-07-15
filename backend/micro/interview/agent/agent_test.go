package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/hub"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	infraEtcd "github.com/yzletter/go-postery/backend/infra/cache/etcd"
	infraLLM "github.com/yzletter/go-postery/backend/infra/llm"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/mcp"
	"github.com/yzletter/go-postery/backend/ports"
)

func TestAgent(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"/")
	fmt.Println(InterviewServiceConf)

	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen) // 初始化千问大模型

	// 测试文档加载器
	loader := ports.NewMyDocLoader()
	text, err := loader.LoadFile(ctx, "test_data/resume.pdf")
	if err != nil {
		fmt.Println("文件加载失败")
		t.Failed()
	}
	fmt.Printf("文档加载成功：\n%s\n", text)

	jdAnalyzer := NewJDAnalyzerAgent(QwenLLMModel)
	resumeMatcher := NewResumeMatcherAgent(QwenLLMModel)

	jdText := `
	后端开发工程师(Go方向)
	要求:
	- 熟练掌握Go 语言，3年以上后端开发经验熟悉 MySQL、Redis、消息队列等常用中间件
	- 有微服务架构设计和开发经验
	- 加分项:有大模型应用开发经验
`

	jdAnalysis, err := jdAnalyzer.Analyze(ctx, jdText)
	jdAnalysisJSON, _ := json.MarshalIndent(jdAnalysis, "", "  ") // 将结构体序列化为带缩进的 JSON
	if err != nil {
		fmt.Println("JD 分析失败")
		t.Failed()
	}
	fmt.Printf("JD 分析结果：\n%+v\n", string(jdAnalysisJSON))

	resume := domain.Resume{RawText: text}
	matchResult, err := resumeMatcher.Match(ctx, jdAnalysis, resume)
	matchJSON, _ := json.MarshalIndent(matchResult, "", "  ") // 将结构体序列化为带缩进的 JSON
	if err != nil {
		fmt.Println("简历匹配失败")
		t.Failed()
	}

	fmt.Printf("简历匹配结果：\n%+v\n", string(matchJSON))
}
func TestResumeAgent(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"/")
	fmt.Println(InterviewServiceConf)

	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen) // 初始化千问大模型

	jdAnalyzer := NewJDAnalyzerAgent(QwenLLMModel)
	resumeMatcher := NewResumeMatcherAgent(QwenLLMModel)

	jdText := `
岗位名称：Go 后端开发工程师\n公司：示例科技\n岗位职责：负责核心业务后端服务开发，参与微服务架构设计与性能优化，建设高可用、高并发的服务系统；负责 MySQL、Redis、消息队列等中间件的使用和优化；参与系统监控、链路追踪和故障排查。\n任职要求：熟悉 Go 语言，理解 goroutine、channel、context、并发安全等机制；熟悉 Gin 或 Kratos 等 Go Web 框架；熟悉 MySQL 索引、事务、锁机制和慢查询优化；熟悉 Redis 常见数据结构、缓存穿透、缓存击穿、缓存雪崩等问题；了解微服务、gRPC、服务注册发现、限流、熔断等机制。\n加分项：有高并发系统经验；熟悉 Kubernetes、Docker；熟悉 OpenTelemetry、Prometheus、Jaeger；有 RAG、LLM 工程化经验。
`
	resumeText := `
候选人姓名：张三\n工作经验：3 年 Go 后端开发经验。\n技术栈：Go、Gin、gRPC、MySQL、Redis、Kafka、Docker、Prometheus。\n项目经历：\n1. 订单系统重构：负责订单服务拆分，将单体服务拆分为订单、支付、库存等微服务，使用 gRPC 通信，接入 etcd 做服务发现。优化 MySQL 慢查询，将部分接口耗时从 800ms 降低到 120ms。\n2. 缓存系统建设：使用 Redis 缓存热点商品数据，解决缓存穿透和缓存击穿问题，使用布隆过滤器和互斥锁方案提升系统稳定性。\n3. 监控告警平台：接入 Prometheus 和 Jaeger，完善接口耗时、错误率、慢 SQL、服务依赖链路监控。\n个人优势：熟悉 Go 并发编程，能独立排查线上问题，了解微服务治理和性能优化。\n不足：Kubernetes 使用经验较少，主要停留在部署和基础排查层面。
`

	jdAnalysis, err := jdAnalyzer.Analyze(ctx, jdText)
	jdAnalysisJSON, _ := json.MarshalIndent(jdAnalysis, "", "  ") // 将结构体序列化为带缩进的 JSON
	if err != nil {
		fmt.Println("JD 分析失败")
		t.Failed()
	}
	fmt.Printf("JD 分析结果：\n%+v\n", string(jdAnalysisJSON))

	resume := domain.Resume{RawText: resumeText}
	matchResult, err := resumeMatcher.Match(ctx, jdAnalysis, resume)
	matchJSON, _ := json.MarshalIndent(matchResult, "", "  ") // 将结构体序列化为带缩进的 JSON
	if err != nil {
		fmt.Println("简历匹配失败")
		t.Failed()
	}

	fmt.Printf("简历匹配结果：\n%+v\n", string(matchJSON))
}

func TestQuestionPlanner(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"/")
	fmt.Println(InterviewServiceConf)

	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen) // 初始化千问大模型

	questionPlanner := NewQuestionPlannerAgent(QwenLLMModel)
	jd := domain.JDAnalysis{
		Position: "Go 后端开发工程师",
		RequiredSkills: []domain.Skill{{
			Name:       "Go",
			Category:   "language",
			Importance: "must",
		}},
		KeyTopics: []string{"Go并发", "微服务", "MySQL"},
	}

	match := domain.ResumeMatchResult{
		OverallScore: 72,
		Strengths:    []string{"Go基础扎实"},
		Weaknesses:   []string{"缺少大模型经验"},
		FocusAreas:   []string{"Go并发", "MySQL优化"},
	}

	plan, err := questionPlanner.PlanDirections(ctx, jd, match, "")
	if err != nil {
		fmt.Println(err)
		t.Failed()
		return
	}
	fmt.Printf("共输出 %d 个方向\n", len(plan.Directions))
	for _, direction := range plan.Directions {
		fmt.Println(direction)
	}
}

func TestEvaluator(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"/")
	fmt.Println(InterviewServiceConf)

	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen) // 初始化千问大模型

	evaluator := NewEvaluatorAgent(QwenLLMModel)

	state := &domain.InterviewState{
		SessionID: 1,
		QAHistory: []domain.QAPair{{
			Question: domain.PlannedQuestion{
				Content:    "请介绍 Go 的 GMP 模型",
				Type:       "basic",
				Difficulty: "medium",
			},
			UserAnswer: "G 是 Goroutine，其他不知道",
			Score:      10,
		}},
		CandidateProfile: "",
	}
	report, err := evaluator.Evaluate(ctx, state, "Go后端开发", "张三", false)
	if err != nil {
		fmt.Println(err)
		t.Failed()
		return
	}

	fmt.Println(report)
}

// 测试复习规划
func TestReview(t *testing.T) {
	ctx := context.Background()
	// 读取配置
	etcdClient := infraEtcd.Init([]string{hub.LocalETCDEndpoint})
	InterviewServiceConf := conf.LoadInterviewServiceConfig(ctx, etcdClient, manager.InterviewService+"/")
	fmt.Println(InterviewServiceConf)

	QwenLLMModel := infraLLM.NewQwenLLMModel(ctx, InterviewServiceConf.Qwen) // 初始化千问大模型

	evaluator := NewEvaluatorAgent(QwenLLMModel)

	state := &domain.InterviewState{
		SessionID: 1,
		QAHistory: []domain.QAPair{{
			Question: domain.PlannedQuestion{
				Content:    "请介绍 Go 的 GMP 模型",
				Type:       "basic",
				Difficulty: "medium",
			},
			UserAnswer: "G 是 Goroutine，其他不知道",
			Score:      10,
		}},
		CandidateProfile: "",
	}
	report, err := evaluator.Evaluate(ctx, state, "Go后端开发", "张三", false)
	if err != nil {
		fmt.Println(err)
		t.Failed()
		return
	}

	review := NewReviewPlannerAgent(QwenLLMModel)
	githubSearcher, err := mcp.NewGitHubSearcher(InterviewServiceConf.Github.Token)
	if err != nil {
		fmt.Println("mcp 初始化失败")
	}
	review.SetGitHubSearcher(githubSearcher)
	plan, err := review.Plan(ctx, report)
	if err != nil {
		fmt.Println(err)
		t.Failed()
		return
	}

	fmt.Println(plan)
}

//go test -v ./backend/micro/interview/agent -run=^TestAgent -count=1
//go test -v ./backend/micro/interview/agent -run=^TestResumeAgent -count=1
//go test -v ./backend/micro/interview/agent -run=^TestQuestionPlanner -count=1
//go test -v ./backend/micro/interview/agent -run=^TestEvaluator -count=1
//go test -v ./backend/micro/interview/agent -run=^TestReview -count=1
