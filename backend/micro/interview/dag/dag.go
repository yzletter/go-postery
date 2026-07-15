package dag

import (
	"context"
	"log/slog"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/dag/node"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
	"github.com/yzletter/go-postery/backend/micro/interview/repository"
	"github.com/yzletter/go-postery/backend/ports"
)

// Orchestrator 建图器
type Orchestrator struct {
	// Agent
	JDAnalyzerAgent      *agent.JDAnalyzerAgent
	ResumeMatcherAgent   *agent.ResumeMatcherAgent
	QuestionPlannerAgent *agent.QuestionPlannerAgent
	InterviewerAgent     *agent.InterviewerAgent
	EvaluatorAgent       *agent.EvaluatorAgent
	ReviewPlannerAgent   *agent.ReviewPlannerAgent

	// 记忆系统
	ShortTermMemory *memory.ShortTermMemory
	LongTermMemory  *memory.LongTermMemory

	// RAG 多路召回
	MultiRetriever *rag.MultiRetriever // 多路召回
	MilvusStore    *rag.MilvusStore    // Milvus 向量存储（支持按用户过滤）
	BM25Manager    *rag.BM25Manager    // BM25 按用户管理
	Reranker       rag.RerankStrategy  // 重排策略（LLM / cross-encoder / none，可切换）

	// 其他
	Callbacks node.FrontendCallbacks
	idGen     ports.IDGenerator
}

func NewOrchestrator(model model.ToolCallingChatModel, repo repository.InterviewRepository, milvusStore *rag.MilvusStore, bm25Manager *rag.BM25Manager, reranker rag.RerankStrategy, idGen ports.IDGenerator) *Orchestrator {
	return &Orchestrator{
		JDAnalyzerAgent:      agent.NewJDAnalyzerAgent(model),
		ResumeMatcherAgent:   agent.NewResumeMatcherAgent(model),
		QuestionPlannerAgent: agent.NewQuestionPlannerAgent(model),
		InterviewerAgent:     agent.NewInterviewerAgent(model),
		EvaluatorAgent:       agent.NewEvaluatorAgent(model),
		ReviewPlannerAgent:   agent.NewReviewPlannerAgent(model),
		ShortTermMemory:      memory.NewShortTermMemory(20),
		LongTermMemory:       memory.NewLongTermMemory(repo),
		MilvusStore:          milvusStore,
		BM25Manager:          bm25Manager,
		Reranker:             reranker,
		idGen:                idGen,
	}
}

// CompilePrepareGraph 建图
func (orchestrator *Orchestrator) CompilePrepareGraph(ctx context.Context) (compose.Runnable[*node.RunState, *node.RunState], error) {
	// Node Builder
	JDAnalyzerNodeBuilder := node.NewJDAnalyzerNodeBuilder(orchestrator.JDAnalyzerAgent, orchestrator.Callbacks, orchestrator.LongTermMemory)
	ResumeMatchNodeBuilder := node.NewResumeMatchNodeBuilder(orchestrator.ResumeMatcherAgent, orchestrator.Callbacks, orchestrator.LongTermMemory)
	QuestionPlanNodeBuilder := node.NewQuestionPlanNodeBuilder(orchestrator.QuestionPlannerAgent, orchestrator.Callbacks, orchestrator.LongTermMemory, orchestrator.MilvusStore, orchestrator.BM25Manager, orchestrator.Reranker)

	// 初始化图状态
	graph := compose.NewGraph[*node.RunState, *node.RunState]()

	// 加点
	JDAnalyzerNode := compose.InvokableLambda(JDAnalyzerNodeBuilder.Build)
	if err := graph.AddLambdaNode(node.JDAnalyzerNodeName, JDAnalyzerNode); err != nil {
		slog.Error("add dag node failed", "node", node.JDAnalyzerNodeName, "error", err)
		return nil, err
	}
	ResumeMatchNode := compose.InvokableLambda(ResumeMatchNodeBuilder.Build)
	if err := graph.AddLambdaNode(node.ResumeMatchNodeName, ResumeMatchNode); err != nil {
		slog.Error("add dag node failed", "node", node.ResumeMatchNodeName, "error", err)
		return nil, err
	}
	QuestionPlanNode := compose.InvokableLambda(QuestionPlanNodeBuilder.Build)
	if err := graph.AddLambdaNode(node.QuestionPlanNodeName, QuestionPlanNode); err != nil {
		slog.Error("add dag node failed", "node", node.QuestionPlanNodeName, "error", err)
		return nil, err
	}

	// 加边
	// START -> JDAnalyzerNode -> ResumeMatchNode -> QuestionPlanNode -> END
	if err := graph.AddEdge(compose.START, node.JDAnalyzerNodeName); err != nil {
		slog.Error("add dag edge failed", "from", compose.START, "to", node.JDAnalyzerNodeName, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.JDAnalyzerNodeName, node.ResumeMatchNodeName); err != nil {
		slog.Error("add dag edge failed", "from", node.JDAnalyzerNodeName, "to", node.ResumeMatchNodeName, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.ResumeMatchNodeName, node.QuestionPlanNodeName); err != nil {
		slog.Error("add dag edge failed", "from", node.ResumeMatchNodeName, "to", node.QuestionPlanNodeName, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.QuestionPlanNodeName, compose.END); err != nil {
		slog.Error("add dag edge failed", "from", node.QuestionPlanNodeName, "to", compose.END, "error", err)
		return nil, err
	}

	// 编译图
	runnable, err := graph.Compile(ctx)
	if err != nil {
		slog.Error("compile dag graph failed", "error", err)
		return nil, err
	}

	// 返回
	return runnable, nil
}

func (orchestrator *Orchestrator) CompileInterViewGraph(ctx context.Context) (compose.Runnable[*node.RunState, *node.RunState], error) {
	// Node Builder
	InterviewerNodeBuilder := node.NewInterviewerNodeBuilder(orchestrator.InterviewerAgent, orchestrator.QuestionPlannerAgent, orchestrator.Callbacks, orchestrator.LongTermMemory)

	// 初始化图状态
	graph := compose.NewGraph[*node.RunState, *node.RunState]()

	InitNode := compose.InvokableLambda(InterviewerNodeBuilder.BuildInitNode)
	if err := graph.AddLambdaNode(node.InterviewInitNode, InitNode); err != nil {
		slog.Error("add dag node failed", "node", node.InterviewInitNode, "error", err)
		return nil, err
	}
	QuestionNode := compose.InvokableLambda(InterviewerNodeBuilder.BuildQuestionNode)
	if err := graph.AddLambdaNode(node.InterviewQuestionNode, QuestionNode); err != nil {
		slog.Error("add dag node failed", "node", node.InterviewQuestionNode, "error", err)
		return nil, err
	}
	AnswerNode := compose.InvokableLambda(InterviewerNodeBuilder.BuildAnswerNode)
	if err := graph.AddLambdaNode(node.InterviewAnswerNode, AnswerNode); err != nil {
		slog.Error("add dag node failed", "node", node.InterviewAnswerNode, "error", err)
		return nil, err
	}
	FollowUpNode := compose.InvokableLambda(InterviewerNodeBuilder.BuildFollowUpNode)
	if err := graph.AddLambdaNode(node.InterviewFollowUpNode, FollowUpNode); err != nil {
		slog.Error("add dag node failed", "node", node.InterviewFollowUpNode, "error", err)
		return nil, err
	}
	UpdateWeakPointsNode := compose.InvokableLambda(InterviewerNodeBuilder.BuildUpdateWPNode)
	if err := graph.AddLambdaNode(node.InterviewUpdateWeakPointsNode, UpdateWeakPointsNode); err != nil {
		slog.Error("add dag node failed", "node", node.InterviewUpdateWeakPointsNode, "error", err)
		return nil, err
	}

	// 加边
	// START -> InterviewInitNode -> InterviewQuestionNode
	if err := graph.AddEdge(compose.START, node.InterviewInitNode); err != nil {
		slog.Error("add dag edge failed", "from", compose.START, "to", node.InterviewInitNode, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.InterviewInitNode, node.InterviewQuestionNode); err != nil {
		slog.Error("add dag edge failed", "from", node.InterviewInitNode, "to", node.InterviewQuestionNode, "error", err)
		return nil, err
	}

	// InterviewQuestionNode 		->  InterviewAnswerNode
	//								->	InterviewFollowUpNode
	//								->	compose.END
	questionNodeBranch := compose.NewGraphBranch[*node.RunState](
		func(ctx context.Context, in *node.RunState) (endNode string, err error) {
			switch in.InterviewState.Phase {
			// 题目问完, 退出图
			case domain.PhaseCompleted:
				return compose.END, nil
			// 等待回答，退出图
			case domain.PhaseWaitingAnswer:
				return compose.END, nil
			// 处理回答
			case domain.PhaseAnswerComing:
				return node.InterviewAnswerNode, nil
			// 处理追问回答
			case domain.PhaseFollowUpComing:
				return node.InterviewFollowUpNode, nil
			default:
				return compose.END, node.ErrInvalidPhaseChange
			}
		},
		map[string]bool{
			compose.END:                true,
			node.InterviewAnswerNode:   true,
			node.InterviewFollowUpNode: true,
		})
	if err := graph.AddBranch(node.InterviewQuestionNode, questionNodeBranch); err != nil {
		slog.Error("add branch edge failed", "node", node.InterviewQuestionNode, "error", err)
		return nil, err
	}

	// InterviewAnswerNode 		->  InterviewFollowUpNode
	//							->	InterviewUpdateWeakPointsNode
	//							->	compose.END
	answerNodeBranch := compose.NewGraphBranch[*node.RunState](
		func(ctx context.Context, in *node.RunState) (endNode string, err error) {
			switch in.InterviewState.Phase {
			// 用户主动退出
			case domain.PhaseUserQuit:
				return compose.END, nil
			// 等待追问回答, 退出图
			case domain.PhaseWaitingFollowUp:
				return compose.END, nil
			// 无需追问
			case domain.PhaseAnswerDone:
				return node.InterviewUpdateWeakPointsNode, nil
			// 处理追问回答
			case domain.PhaseFollowUpComing:
				return node.InterviewFollowUpNode, nil
			default:
				return compose.END, node.ErrInvalidPhaseChange
			}
		},
		map[string]bool{
			compose.END:                        true,
			node.InterviewFollowUpNode:         true,
			node.InterviewUpdateWeakPointsNode: true,
		})
	if err := graph.AddBranch(node.InterviewAnswerNode, answerNodeBranch); err != nil {
		slog.Error("add branch edge failed", "node", node.InterviewAnswerNode, "error", err)
		return nil, err
	}

	// InterviewFollowUpNode 	->  InterviewUpdateWeakPointsNode
	//							->	compose.END
	followUpNodeBranch := compose.NewGraphBranch[*node.RunState](
		func(ctx context.Context, in *node.RunState) (endNode string, err error) {
			switch in.InterviewState.Phase {
			// 用户主动退出
			case domain.PhaseUserQuit:
				return compose.END, nil
			// 追问回答完毕, 更新薄弱点
			case domain.PhaseFollowUpDone:
				return node.InterviewUpdateWeakPointsNode, nil
			default:
				return compose.END, node.ErrInvalidPhaseChange
			}
		},
		map[string]bool{
			compose.END:                        true,
			node.InterviewUpdateWeakPointsNode: true,
		})
	if err := graph.AddBranch(node.InterviewFollowUpNode, followUpNodeBranch); err != nil {
		slog.Error("add branch edge failed", "node", node.InterviewFollowUpNode, "error", err)
		return nil, err
	}

	// InterviewUpdateWeakPointsNode -> InterviewQuestionNode
	if err := graph.AddEdge(node.InterviewUpdateWeakPointsNode, node.InterviewQuestionNode); err != nil {
		slog.Error("add dag edge failed", "from", node.InterviewUpdateWeakPointsNode, "to", node.InterviewQuestionNode, "error", err)
		return nil, err
	}

	// 编译图
	runnable, err := graph.Compile(ctx)
	if err != nil {
		slog.Error("compile dag graph failed", "error", err)
		return nil, err
	}

	// 返回
	return runnable, nil
}

func (orchestrator *Orchestrator) CompileEvaluationGraph(ctx context.Context) (compose.Runnable[*node.RunState, *node.RunState], error) {
	// Node Builder
	WeakReviewNodeBuilder := node.NewWeakReviewNodeBuilder(orchestrator.Callbacks, orchestrator.LongTermMemory, orchestrator.MilvusStore, orchestrator.BM25Manager, orchestrator.Reranker)
	EvaluationNodeBuilder := node.NewEvaluationNodeBuilder(orchestrator.EvaluatorAgent, orchestrator.Callbacks, orchestrator.LongTermMemory)
	ReviewPlanNodeBuilder := node.NewReviewPlannerNodeBuilder(orchestrator.ReviewPlannerAgent, orchestrator.Callbacks, orchestrator.LongTermMemory, orchestrator.idGen)

	// 初始化图状态
	graph := compose.NewGraph[*node.RunState, *node.RunState]()

	WeakReviewNode := compose.InvokableLambda(WeakReviewNodeBuilder.Build)
	if err := graph.AddLambdaNode(node.WeakReviewNodeName, WeakReviewNode); err != nil {
		slog.Error("add dag node failed", "node", node.WeakReviewNodeName, "error", err)
		return nil, err
	}
	EvaluationNode := compose.InvokableLambda(EvaluationNodeBuilder.Build)
	if err := graph.AddLambdaNode(node.EvaluationNodeName, EvaluationNode); err != nil {
		slog.Error("add dag node failed", "node", node.EvaluationNodeName, "error", err)
		return nil, err
	}
	ReviewPlanNode := compose.InvokableLambda(ReviewPlanNodeBuilder.Build)
	if err := graph.AddLambdaNode(node.ReviewPlannerNodeName, ReviewPlanNode); err != nil {
		slog.Error("add dag node failed", "node", node.ReviewPlannerNodeName, "error", err)
		return nil, err
	}

	// 加边
	// START -> WeakReviewNode -> EvaluationNode -> ReviewPlanNode -> END
	if err := graph.AddEdge(compose.START, node.WeakReviewNodeName); err != nil {
		slog.Error("add dag edge failed", "from", compose.START, "to", node.JDAnalyzerNodeName, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.WeakReviewNodeName, node.EvaluationNodeName); err != nil {
		slog.Error("add dag edge failed", "from", node.WeakReviewNodeName, "to", node.EvaluationNodeName, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.EvaluationNodeName, node.ReviewPlannerNodeName); err != nil {
		slog.Error("add dag edge failed", "from", node.EvaluationNodeName, "to", node.ReviewPlannerNodeName, "error", err)
		return nil, err
	}
	if err := graph.AddEdge(node.ReviewPlannerNodeName, compose.END); err != nil {
		slog.Error("add dag edge failed", "from", node.ReviewPlannerNodeName, "to", compose.END, "error", err)
		return nil, err
	}

	// 编译图
	runnable, err := graph.Compile(ctx)
	if err != nil {
		slog.Error("compile dag graph failed", "error", err)
		return nil, err
	}

	// 返回
	return runnable, nil
}
