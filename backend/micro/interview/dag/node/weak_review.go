package node

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/micro/interview/domain"
	"github.com/yzletter/go-postery/backend/micro/interview/memory"
	"github.com/yzletter/go-postery/backend/micro/interview/rag"
)

type WeakReviewNodeBuilder struct {
	Callbacks      FrontendCallbacks
	LongTermMemory *memory.LongTermMemory
	MilvusStore    *rag.MilvusStore   // Milvus 向量存储（支持按用户过滤）
	BM25Manager    *rag.BM25Manager   // BM25 按用户管理
	Reranker       rag.RerankStrategy // 重排策略（LLM / cross-encoder / none，可切换）
}

func NewWeakReviewNodeBuilder(callbacks FrontendCallbacks, longTermMemory *memory.LongTermMemory, milvusStore *rag.MilvusStore, bm25Manager *rag.BM25Manager, reranker rag.RerankStrategy) *WeakReviewNodeBuilder {
	return &WeakReviewNodeBuilder{
		Callbacks:      callbacks,
		LongTermMemory: longTermMemory,
		MilvusStore:    milvusStore,
		BM25Manager:    bm25Manager,
		Reranker:       reranker,
	}
}

func (builder *WeakReviewNodeBuilder) Build(ctx context.Context, input *RunState) (*RunState, error) {
	if len(input.InterviewState.QAHistory) == 0 {
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageTerminated, "面试未作答即终止，不生成评估报告。")
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}
		return input, nil
	}

	if input.UserTerminated {
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageTerminated, fmt.Sprintf("用户主动终止面试（已完成 %d/%d 题）", len(input.InterviewState.QAHistory), input.InterviewState.TotalQuestions))
	}

	// 遍历历史回答
	if len(input.InterviewState.QAHistory) == 0 {
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageWeakReviewStart, "正在检查低分题目...")
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageWeakReviewDone, "暂无答题记录，跳过低分题目巩固")
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}
		return input, nil
	}
	weakQAPairs := make([]domain.QAPair, 0)
	for _, pair := range input.InterviewState.QAHistory {
		if pair.Score < 60 {
			weakQAPairs = append(weakQAPairs, pair)
		}
	}
	if len(weakQAPairs) == 0 {
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageWeakReviewStart, "正在检查低分题目...")
		builder.Callbacks.OnStageChange(ctx, input.UserID, StageWeakReviewDone, "没有低分题目需要巩固")
		if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
			slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
			return input, err
		}
		return input, nil
	}

	builder.Callbacks.OnStageChange(ctx, input.UserID, StageWeakReviewStart, fmt.Sprintf("正在对 %d 道低分题目进行巩固...", len(weakQAPairs)))

	for idx, qa := range weakQAPairs {
		var text string

		// LLM 题目直接返回答案
		if qa.Question.Source == domain.QuestionSourceLLM {
			if qa.Question.Reference != "" {
				text = fmt.Sprintf("**低分题目巩固 %d/%d**\n\n**题目：** %s\n\n**你的得分：** %.0f\n\n**参考答案：**\n%s", idx+1, len(weakQAPairs), qa.Question.Content, qa.Score, qa.Question.Reference)
			}
		} else if qa.Question.Source != "" {
			// 检索的题目
			refAnswer := qa.Question.Reference
			if refAnswer == "" && (builder.MilvusStore != nil || builder.BM25Manager != nil) { // 答案为空, 进行检索兜底
				query := qa.Question.Content
				var docs []*schema.Document
				mp := make(map[string]struct{})

				// BM25
				if builder.BM25Manager != nil {
					bm25Docs, err := builder.BM25Manager.Retrieve(ctx, input.UserID, query)
					if err != nil {
						slog.Warn("retrieve weak review by bm25 failed", "user_id", input.UserID, "session_id", input.ID, "query", query, "error", err)
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
					milvusDocs, err := builder.MilvusStore.RetrieveByUser(ctx, input.UserID, query)
					if err != nil {
						slog.Warn("retrieve weak review by milvus failed", "user_id", input.UserID, "session_id", input.ID, "query", query, "error", err)
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

				// 重排
				if len(docs) > 0 {
					selectedDoc := docs[0]
					if builder.Reranker != nil {
						rerankedDocs, err := builder.Reranker.Rerank(ctx, query, docs)
						if err != nil {
							slog.Warn("rerank weak review docs failed", "user_id", input.UserID, "session_id", input.ID, "query", query, "error", err)
						}

						if rerankedDocs != nil && len(rerankedDocs) > 0 {
							// 取当前方向 Top1
							selectedDoc = rerankedDocs[0]
						}
					}

					content := selectedDoc.Content
					if aIdx := strings.Index(content, "\n参考答案："); aIdx >= 0 {
						refAnswer = strings.TrimSpace(content[aIdx+len("\n参考答案："):])
					}
				}
			}

			if refAnswer != "" {
				text = fmt.Sprintf("**低分题目巩固 %d/%d**\n\n**题目：** %s\n\n**你的得分：** %.0f\n\n**题库参考答案：**\n%s",
					idx+1, len(weakQAPairs), qa.Question.Content, qa.Score, refAnswer)
			}
		}

		// 返回前端
		if text != "" {
			builder.Callbacks.OnQuestion(ctx, input.UserID, 0, text)
		}
	}

	builder.Callbacks.OnStageChange(ctx, input.UserID, StageWeakReviewDone, "低分题目巩固完成")

	// 更新会话信息
	if err := builder.LongTermMemory.UpsertSession(ctx, input.UserID, input.ID, input); err != nil {
		slog.Error("upsert session failed", "user_id", input.UserID, "session_id", input.ID, "error", err)
		return input, err
	}

	return input, nil
}
