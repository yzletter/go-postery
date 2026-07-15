package node

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/agent"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
	"github.com/yzletter/go-postery/backend/micro/interview/model"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
)

type QuestionPlanNodeBuilder struct {
	QuestionPlannerAgent *agent.QuestionPlannerAgent
	Callbacks            FrontendCallbacks
	LongTermMemory       *memory.LongTermMemory
	MilvusStore          *rag.MilvusStore   // Milvus 向量存储（支持按用户过滤）
	BM25Manager          *rag.BM25Manager   // BM25 按用户管理
	Reranker             rag.RerankStrategy // 重排策略（LLM / cross-encoder / none，可切换）
}

func NewQuestionPlanNodeBuilder(agent *agent.QuestionPlannerAgent, callbacks FrontendCallbacks,
	longTermMemory *memory.LongTermMemory, milvusStore *rag.MilvusStore, bm25Manager *rag.BM25Manager, reranker rag.RerankStrategy) *QuestionPlanNodeBuilder {
	return &QuestionPlanNodeBuilder{
		QuestionPlannerAgent: agent,
		Callbacks:            callbacks,
		LongTermMemory:       longTermMemory,
		MilvusStore:          milvusStore,
		BM25Manager:          bm25Manager,
		Reranker:             reranker,
	}
}

func (builder *QuestionPlanNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
	// 出题规划开始
	builder.Callbacks.OnStageChange(ctx, input.UserID, StageQuestionPlan, "正在进行出题规划 ... ")

	// 获取薄弱点
	var weakPointsText string
	weakPoints := builder.LongTermMemory.GetWeakPoints(ctx, input.UserID)
	// 按 JD 把薄弱点过滤并转为文本
	filteredWeakPoints := filterWeakPointsWithJD(input.JDAnalysis, weakPoints)
	if len(filteredWeakPoints) > 0 {
		weakPointsText = strings.Join(filteredWeakPoints, "\n")
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageMemoryLoaded, fmt.Sprintf("已加载 %d 个与当前 JD 相关的历史薄弱点，将针对性出题", len(filteredWeakPoints)))
	}
	
	// 出题方向
	directionPlan, err := builder.QuestionPlannerAgent.PlanDirections(ctx, *input.JDAnalysis, *input.ResumeMatchResult, weakPointsText)
	if err != nil {
		slog.Error("plan question directions failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	// 已匹配的问题
	matchedQuestions := make([]domain.PlannedQuestion, 0)
	matchedCnt := 0
	// 未匹配的方向
	unmatchedDirections := make([]domain.QuestionDirection, 0)

	// 检索题目
	hasRAG := builder.BM25Manager != nil || builder.MilvusStore != nil
	if hasRAG {
		for _, direction := range directionPlan.Directions {
			// 只检索 basic 类题目，其余由 LLM 生成
			if direction.Type != domain.QuestionTypeBasic {
				unmatchedDirections = append(unmatchedDirections, direction)
				continue
			}

			var docs []*schema.Document
			mp := make(map[string]struct{})

			// BM25
			if builder.BM25Manager != nil {
				bm25Docs, err := builder.BM25Manager.Retrieve(ctx, input.UserID, direction.SearchQuery)
				if err != nil {
					slog.Warn("retrieve question by bm25 failed", "user_id", input.UserID, "session_id", input.ID, "query", direction.SearchQuery, "error", err)
				} else {
					// 去重
					for _, doc := range bm25Docs {
						if _, exist := mp[doc.ID]; !exist {
							docs = append(docs, doc)
							mp[doc.ID] = struct{}{}
						}
					}
				}
			}

			// Milvus
			if builder.MilvusStore != nil {
				milvusDocs, err := builder.MilvusStore.RetrieveByUser(ctx, input.UserID, direction.SearchQuery)
				if err != nil {
					slog.Warn("retrieve question by milvus failed", "user_id", input.UserID, "session_id", input.ID, "query", direction.SearchQuery, "error", err)
				} else {
					// 去重
					for _, doc := range milvusDocs {
						if _, exist := mp[doc.ID]; !exist {
							docs = append(docs, doc)
							mp[doc.ID] = struct{}{}
						}
					}
				}
			}

			if len(docs) == 0 {
				slog.Info("question not matched", "user_id", input.UserID, "session_id", input.ID, "topic", direction.Topic, "query", direction.SearchQuery)
				// 将该方向放入未匹配方向交由 LLM 生成题目
				unmatchedDirections = append(unmatchedDirections, direction)
				continue
			}

			selectedDoc := docs[0]

			// 重排
			if builder.Reranker != nil {
				rerankedDocs, err := builder.Reranker.Rerank(ctx, direction.SearchQuery, docs)
				if err != nil {
					slog.Warn("rerank question docs failed", "user_id", input.UserID, "session_id", input.ID, "query", direction.SearchQuery, "error", err)
				}

				if rerankedDocs != nil && len(rerankedDocs) > 0 {
					// 取当前方向 Top1
					selectedDoc = rerankedDocs[0]
				}
			}

			// 直接用题库原题构建题目，不经过 LLM
			questionContent := selectedDoc.Content
			reference := ""
			if idx := strings.Index(questionContent, "\n参考答案："); idx >= 0 {
				reference = strings.TrimSpace(questionContent[idx+len("\n参考答案："):])
				questionContent = strings.TrimSpace(questionContent[:idx])
			} else if idx := strings.Index(questionContent, "\n参考答案:"); idx >= 0 {
				reference = strings.TrimSpace(questionContent[idx+len("\n参考答案:"):])
				questionContent = strings.TrimSpace(questionContent[:idx])
			}

			matchedQuestions = append(matchedQuestions, domain.PlannedQuestion{
				Content:    questionContent,
				Type:       direction.Type,
				Difficulty: direction.Difficulty,
				Skills:     direction.Skills,
				FollowUps:  []string{},
				Reference:  reference,
				Source:     selectedDoc.ID,
			})

			matchedCnt++
		}
	} else {
		// 没有检索功能
		unmatchedDirections = append(unmatchedDirections, directionPlan.Directions...)
	}

	// 未匹配的 Direction 进行出题
	llmQuestions := make([]domain.PlannedQuestion, 0)
	if len(unmatchedDirections) > 0 {
		unmatchedPlan := domain.QuestionDirectionPlan{Directions: unmatchedDirections}
		emptyDocs := make([]string, len(unmatchedDirections)) // 全部无匹配
		assembledQuestions, err := builder.QuestionPlannerAgent.AssembleQuestion(ctx, input.JDAnalysis, input.ResumeMatchResult, unmatchedPlan, emptyDocs)
		if err != nil {
			slog.Error("assemble questions failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}
		llmQuestions = assembledQuestions.Questions
	}

	// 合并题目
	questions := append(matchedQuestions, llmQuestions...)
	for i := range questions { // 编号
		questions[i].ID = int64(i + 1)
	}

	// 统计分布
	var basicCount, expCount, designCount int
	for _, q := range questions {
		switch q.Type {
		case domain.QuestionTypeBasic:
			basicCount++
		case domain.QuestionTypeExperience:
			expCount++
		case domain.QuestionTypeDesign:
			designCount++
		}
	}

	plan := &domain.QuestionPlan{
		TotalQuestions: len(questions),
		Distribution:   domain.QuestionDistribution{Basic: basicCount, Experience: expCount, Design: designCount},
		Questions:      questions,
	}

	// 更新上下文
	input.QuestionPlan = plan
	input.Status = domain.StatusPlanned

	// 出题规划结束
	builder.Callbacks.OnStageChange(ctx, input.UserID, StageQuestionPlanDone, "")

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	return input, nil
}

// 简单按照 JD 过滤 WeakPoint, 检查 JD 的需要技能、加分技能、关键词与 WeakPoint 是否相互包含
func filterWeakPointsWithJD(jd *domain.JDAnalysis, weakPoints []model.WeakPoint) []string {
	var skills []string
	for _, s := range jd.RequiredSkills {
		skills = append(skills, strings.ToLower(s.Name))
	}
	for _, s := range jd.PreferredSkills {
		skills = append(skills, strings.ToLower(s.Name))
	}
	for _, t := range jd.KeyTopics {
		skills = append(skills, strings.ToLower(t))
	}

	var res []string
	for _, wp := range weakPoints {
		// 转小写
		wt := strings.ToLower(wp.Topic)
		// 检查是否包含
		for _, skill := range skills {
			if strings.Contains(wt, skill) || strings.Contains(skill, wt) {
				res = append(res, fmt.Sprintf("- %s：历史得分 %.0f，被考察 %d 次，答错 %d 次",
					wp.Topic, wp.Score, wp.HitCount, wp.WrongCount))
			}
		}
	}
	if len(res) > 0 {
		return res
	}
	return nil
}
