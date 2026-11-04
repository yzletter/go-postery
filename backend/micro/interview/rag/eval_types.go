package rag

import "time"

// EvalSample 一条评估样本（评估数据集中的一行）
type EvalSample struct {
	ID             string   `json:"id"`               // 样本唯一 ID，如 "eval_001"
	Query          string   `json:"query"`            // 检索查询（模拟 Phase 1 的 SearchQuery）
	RelevantDocIDs []string `json:"relevant_doc_ids"` // 人工标注的相关题目 ID 列表（黄金标准）
	Topic          string   `json:"topic"`            // 技术领域，如 "Go 并发" / "MySQL"
	Difficulty     string   `json:"difficulty"`       // 难度：easy / medium / hard
	Note           string   `json:"note,omitempty"`   // 标注说明（可选）：什么算相关、什么不算
}

// SampleResult 单条样本的评估结果
type SampleResult struct {
	SampleID     string   `json:"sample_id"`
	Query        string   `json:"query"`
	Topic        string   `json:"topic,omitempty"`
	Difficulty   string   `json:"difficulty,omitempty"`
	RetrievedIDs []string `json:"retrieved_ids"` // RAG 实际返回的有序 ID 列表（Rerank 后的顺序）
	RelevantIDs  []string `json:"relevant_ids"`  // 标注的相关 ID
	HitIDs       []string `json:"hit_ids"`       // 命中的交集（按检索结果中出现的顺序）
	RecallAt10   float64  `json:"recall_at_10"`
	RecallAt20   float64  `json:"recall_at_20"`
	MRR          float64  `json:"mrr"`
	FirstHitRank int      `json:"first_hit_rank"` // 第一个命中的排名（1-based，0 表示未命中）
}

// MetricsSummary 指标汇总（用于整体 / 按 topic / 按 difficulty 分组）
type MetricsSummary struct {
	Count      int     `json:"count"`
	RecallAt10 float64 `json:"recall_at_10"`
	RecallAt20 float64 `json:"recall_at_20"`
	MRR        float64 `json:"mrr"`
}

// RAGConfig 评估时的 RAG 配置快照（用于 A/B 对比时回溯当时的参数）
//
// 跑评估时把当前 RAG 流水线的关键参数记到这里，写入报告。
// 比较两份报告时，从这个字段就能看出"两次实验的参数差异在哪"。
type RAGConfig struct {
	EmbeddingModel string  `json:"embedding_model,omitempty"`
	VectorDim      int     `json:"vector_dim,omitempty"`
	VectorTopK     int     `json:"vector_top_k,omitempty"`
	BM25TopK       int     `json:"bm25_top_k,omitempty"`
	BM25K1         float64 `json:"bm25_k1,omitempty"`
	BM25B          float64 `json:"bm25_b,omitempty"`
	RerankerType   string  `json:"reranker_type,omitempty"` // "llm" / "gte-rerank" / "none"
	RerankTopN     int     `json:"rerank_top_n,omitempty"`
	Note           string  `json:"note,omitempty"` // 实验备注，如 "baseline" / "调 BM25 k1 到 1.2"
}

// EvalReport 一次评估运行的完整报告
type EvalReport struct {
	DatasetVersion    string                    `json:"dataset_version"`
	DatasetPath       string                    `json:"dataset_path,omitempty"`
	SampleCount       int                       `json:"sample_count"`
	RunAt             time.Time                 `json:"run_at"`
	Duration          string                    `json:"duration"`
	Config            RAGConfig                 `json:"config"`
	Overall           MetricsSummary            `json:"overall"`
	TopicMetrics      map[string]MetricsSummary `json:"topic_metrics,omitempty"`
	DifficultyMetrics map[string]MetricsSummary `json:"difficulty_metrics,omitempty"`
	SampleResults     []SampleResult            `json:"sample_results"`
	WorstSamples      []SampleResult            `json:"worst_samples,omitempty"` // 按 Recall@10 升序的前 10 条
}
