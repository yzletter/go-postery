package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/cloudwego/eino/schema"
)

// RunEvaluationConfig 评估流水线的运行配置
type RunEvaluationConfig struct {
	Dataset     []EvalSample
	UserID      int64 // 题库按 user 隔离，评估用哪个用户的题库
	MilvusStore *MilvusStore
	BM25Manager *BM25Manager
	Reranker    RerankStrategy // 重排策略（llm / cross-encoder / none，可切换）
	RAGConfig   RAGConfig      // 用于报告快照，便于 A/B 对比
}

// LoadEvalDataset 从 JSON 文件加载评估数据集
func LoadEvalDataset(filePath string) ([]EvalSample, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("eval: read dataset %s: %w", filePath, err)
	}
	var samples []EvalSample
	if err := json.Unmarshal(data, &samples); err != nil {
		return nil, fmt.Errorf("eval: parse dataset: %w", err)
	}
	return samples, nil
}

// RunEvaluation 对评估数据集做全量检索评估，返回完整报告
//
// 整体流程：
//  1. 遍历每条样本，用 sample.Query 跑完整的 RAG 检索（Milvus + BM25 + Rerank）
//  2. 对比检索结果和 sample.RelevantDocIDs，计算 Recall@10 / Recall@20 / MRR
//  3. 汇总成 EvalReport，附带按 topic / difficulty 分组指标与异常样本列表
func RunEvaluation(ctx context.Context, cfg *RunEvaluationConfig) (*EvalReport, error) {
	if len(cfg.Dataset) == 0 {
		return nil, fmt.Errorf("eval: dataset is empty")
	}
	start := time.Now()
	results := make([]SampleResult, 0, len(cfg.Dataset))

	for i, sample := range cfg.Dataset {
		log.Printf("[Eval] (%d/%d) sample=%s query=%q", i+1, len(cfg.Dataset), sample.ID, sample.Query)

		retrievedDocs, err := retrieveForEval(ctx, cfg, sample.Query)
		if err != nil {
			log.Printf("[Eval] sample %s retrieve failed: %v", sample.ID, err)
			continue
		}
		retrievedIDs := docsToIDs(retrievedDocs)

		recall10 := calcRecallAtK(retrievedIDs, sample.RelevantDocIDs, 10)
		recall20 := calcRecallAtK(retrievedIDs, sample.RelevantDocIDs, 20)
		mrr := calcMRR(retrievedIDs, sample.RelevantDocIDs)
		hits, firstHit := calcHits(retrievedIDs, sample.RelevantDocIDs)

		results = append(results, SampleResult{
			SampleID:     sample.ID,
			Query:        sample.Query,
			Topic:        sample.Topic,
			Difficulty:   sample.Difficulty,
			RetrievedIDs: retrievedIDs,
			RelevantIDs:  sample.RelevantDocIDs,
			HitIDs:       hits,
			RecallAt10:   recall10,
			RecallAt20:   recall20,
			MRR:          mrr,
			FirstHitRank: firstHit,
		})
	}

	report := aggregateReport(results, cfg.RAGConfig)
	report.Duration = time.Since(start).String()
	return report, nil
}

// retrieveForEval 评估时的检索流程
//
// 必须和 orchestrator.go Phase 1.5 保持完全一致：
// BM25 → Milvus 双路召回 → ID 去重 → LLM Reranker。
// 任何主流程变更都要同步修改这里，否则评估结果就和线上分布偏离。
func retrieveForEval(ctx context.Context, cfg *RunEvaluationConfig, query string) ([]*schema.Document, error) {
	var docs []*schema.Document
	seen := make(map[string]bool)

	if cfg.BM25Manager != nil {
		bm25Docs, err := cfg.BM25Manager.Retrieve(ctx, cfg.UserID, query)
		if err != nil {
			log.Printf("[Eval] BM25 retrieve err: %v", err)
		}
		for _, d := range bm25Docs {
			if !seen[d.ID] {
				seen[d.ID] = true
				docs = append(docs, d)
			}
		}
	}

	if cfg.MilvusStore != nil {
		milvusDocs, err := cfg.MilvusStore.RetrieveByUser(ctx, cfg.UserID, query)
		if err != nil {
			log.Printf("[Eval] Milvus retrieve err: %v", err)
		}
		for _, d := range milvusDocs {
			if !seen[d.ID] {
				seen[d.ID] = true
				docs = append(docs, d)
			}
		}
	}

	if cfg.Reranker != nil && len(docs) > 1 {
		reranked, err := cfg.Reranker.Rerank(ctx, query, docs)
		if err == nil && len(reranked) > 0 {
			docs = reranked
		}
	}

	return docs, nil
}

func docsToIDs(docs []*schema.Document) []string {
	ids := make([]string, len(docs))
	for i, d := range docs {
		ids[i] = d.ID
	}
	return ids
}

// calcRecallAtK = |Top-K ∩ Relevant| / |Relevant|
//
// 如果 relevantIDs 为空，返回 0（这种样本本身就有问题，应该在标注阶段避免）。
func calcRecallAtK(retrievedIDs, relevantIDs []string, k int) float64 {
	if len(relevantIDs) == 0 {
		return 0
	}
	relevantSet := make(map[string]bool, len(relevantIDs))
	for _, id := range relevantIDs {
		relevantSet[id] = true
	}
	limit := k
	if limit > len(retrievedIDs) {
		limit = len(retrievedIDs)
	}
	hits := 0
	for i := 0; i < limit; i++ {
		if relevantSet[retrievedIDs[i]] {
			hits++
		}
	}
	return float64(hits) / float64(len(relevantIDs))
}

// calcMRR 第一个相关文档的排名倒数（rank 1-based，未命中返回 0）
func calcMRR(retrievedIDs, relevantIDs []string) float64 {
	relevantSet := make(map[string]bool, len(relevantIDs))
	for _, id := range relevantIDs {
		relevantSet[id] = true
	}
	for i, id := range retrievedIDs {
		if relevantSet[id] {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}

// calcHits 返回命中的 ID 集合（按检索结果顺序）和第一个命中的排名（1-based, 0 表示未命中）
func calcHits(retrievedIDs, relevantIDs []string) ([]string, int) {
	relevantSet := make(map[string]bool, len(relevantIDs))
	for _, id := range relevantIDs {
		relevantSet[id] = true
	}
	hits := make([]string, 0)
	firstHit := 0
	for i, id := range retrievedIDs {
		if relevantSet[id] {
			hits = append(hits, id)
			if firstHit == 0 {
				firstHit = i + 1
			}
		}
	}
	return hits, firstHit
}

// aggregateReport 把单样本结果汇总成完整的 EvalReport
//
// 包含：整体指标 / 按 topic 分组 / 按 difficulty 分组 / 异常样本 Top-10。
func aggregateReport(results []SampleResult, ragCfg RAGConfig) *EvalReport {
	report := &EvalReport{
		DatasetVersion:    "v1",
		SampleCount:       len(results),
		RunAt:             time.Now(),
		Config:            ragCfg,
		SampleResults:     results,
		TopicMetrics:      make(map[string]MetricsSummary),
		DifficultyMetrics: make(map[string]MetricsSummary),
	}
	if len(results) == 0 {
		return report
	}

	report.Overall = computeMetricsSummary(results)

	// 按 topic 分组
	topicGroups := make(map[string][]SampleResult)
	for _, r := range results {
		if r.Topic != "" {
			topicGroups[r.Topic] = append(topicGroups[r.Topic], r)
		}
	}
	for topic, group := range topicGroups {
		report.TopicMetrics[topic] = computeMetricsSummary(group)
	}

	// 按 difficulty 分组
	diffGroups := make(map[string][]SampleResult)
	for _, r := range results {
		if r.Difficulty != "" {
			diffGroups[r.Difficulty] = append(diffGroups[r.Difficulty], r)
		}
	}
	for diff, group := range diffGroups {
		report.DifficultyMetrics[diff] = computeMetricsSummary(group)
	}

	// 异常样本：按 Recall@10 升序，再按 MRR 升序，取前 10 条
	sorted := make([]SampleResult, len(results))
	copy(sorted, results)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].RecallAt10 != sorted[j].RecallAt10 {
			return sorted[i].RecallAt10 < sorted[j].RecallAt10
		}
		return sorted[i].MRR < sorted[j].MRR
	})
	limit := 10
	if limit > len(sorted) {
		limit = len(sorted)
	}
	report.WorstSamples = sorted[:limit]

	return report
}

func computeMetricsSummary(results []SampleResult) MetricsSummary {
	if len(results) == 0 {
		return MetricsSummary{}
	}
	var sumR10, sumR20, sumMRR float64
	for _, r := range results {
		sumR10 += r.RecallAt10
		sumR20 += r.RecallAt20
		sumMRR += r.MRR
	}
	n := float64(len(results))
	return MetricsSummary{
		Count:      len(results),
		RecallAt10: roundFloat(sumR10/n, 4),
		RecallAt20: roundFloat(sumR20/n, 4),
		MRR:        roundFloat(sumMRR/n, 4),
	}
}

func roundFloat(v float64, decimals int) float64 {
	shift := math.Pow(10, float64(decimals))
	return math.Round(v*shift) / shift
}
