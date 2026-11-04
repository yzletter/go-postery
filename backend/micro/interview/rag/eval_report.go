package rag

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func SaveReportJSON(report *EvalReport, filePath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write report to %s: %w", filePath, err)
	}
	return nil
}

func SaveReportMarkdown(report *EvalReport, filePath string) error {
	md := RenderReportMarkdown(report)
	return os.WriteFile(filePath, []byte(md), 0644)
}

// RenderReportMarkdown 将 EvalReport 渲染为可读的 Markdown 摘要
// 包含：基本信息、配置快照、整体指标、按 topic 分组、按 difficulty 分组、Worst-10
func RenderReportMarkdown(report *EvalReport) string {
	var b strings.Builder
	renderHeader(&b, report)
	renderConfig(&b, report)
	renderOverall(&b, report)
	renderTopicMetrics(&b, report)
	renderDifficultyMetrics(&b, report)
	renderWorstSamples(&b, report)
	return b.String()
}

func renderHeader(b *strings.Builder, r *EvalReport) {
	b.WriteString("# RAG 离线评估报告\n\n")
	b.WriteString(fmt.Sprintf("- **运行时间**：%s\n", r.RunAt.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("- **数据集版本**：%s\n", r.DatasetVersion))
	if r.DatasetPath != "" {
		b.WriteString(fmt.Sprintf("- **数据集路径**：`%s`\n", r.DatasetPath))
	}
	b.WriteString(fmt.Sprintf("- **样本数**：%d\n", r.SampleCount))
	b.WriteString(fmt.Sprintf("- **耗时**：%s\n\n", r.Duration))
}

func renderConfig(b *strings.Builder, r *EvalReport) {
	b.WriteString("## 1. 配置快照\n\n| 字段 | 值 |\n|------|----|\n")
	c := r.Config
	if c.EmbeddingModel != "" {
		b.WriteString(fmt.Sprintf("| EmbeddingModel | %s |\n", c.EmbeddingModel))
	}
	if c.VectorDim > 0 {
		b.WriteString(fmt.Sprintf("| VectorDim | %d |\n", c.VectorDim))
	}
	if c.VectorTopK > 0 {
		b.WriteString(fmt.Sprintf("| VectorTopK | %d |\n", c.VectorTopK))
	}
	if c.BM25TopK > 0 {
		b.WriteString(fmt.Sprintf("| BM25TopK | %d |\n", c.BM25TopK))
	}
	if c.BM25K1 > 0 {
		b.WriteString(fmt.Sprintf("| BM25 k1 | %.2f |\n", c.BM25K1))
	}
	if c.BM25B > 0 {
		b.WriteString(fmt.Sprintf("| BM25 b | %.2f |\n", c.BM25B))
	}
	if c.RerankerType != "" {
		b.WriteString(fmt.Sprintf("| RerankerType | %s |\n", c.RerankerType))
	}
	if c.RerankTopN > 0 {
		b.WriteString(fmt.Sprintf("| RerankTopN | %d |\n", c.RerankTopN))
	}
	if c.Note != "" {
		b.WriteString(fmt.Sprintf("| Note | %s |\n", c.Note))
	}
	b.WriteString("\n")
}

func renderOverall(b *strings.Builder, r *EvalReport) {
	b.WriteString("## 2. 整体指标\n\n| 指标 | 值 |\n|------|----|\n")
	b.WriteString(fmt.Sprintf("| Recall@10 | %.4f |\n", r.Overall.RecallAt10))
	b.WriteString(fmt.Sprintf("| Recall@20 | %.4f |\n", r.Overall.RecallAt20))
	b.WriteString(fmt.Sprintf("| MRR       | %.4f |\n\n", r.Overall.MRR))
}

func renderTopicMetrics(b *strings.Builder, r *EvalReport) {
	if len(r.TopicMetrics) == 0 {
		return
	}
	b.WriteString("## 3. 按 Topic 分组\n\n| Topic | Count | Recall@10 | Recall@20 | MRR |\n|-------|-------|-----------|-----------|-----|\n")
	topics := make([]string, 0, len(r.TopicMetrics))
	for k := range r.TopicMetrics {
		topics = append(topics, k)
	}
	sort.Strings(topics)
	for _, t := range topics {
		m := r.TopicMetrics[t]
		b.WriteString(fmt.Sprintf("| %s | %d | %.4f | %.4f | %.4f |\n",
			t, m.Count, m.RecallAt10, m.RecallAt20, m.MRR))
	}
	b.WriteString("\n")
}

func renderDifficultyMetrics(b *strings.Builder, r *EvalReport) {
	if len(r.DifficultyMetrics) == 0 {
		return
	}
	b.WriteString("## 4. 按难度分组\n\n| Difficulty | Count | Recall@10 | Recall@20 | MRR |\n|------------|-------|-----------|-----------|-----|\n")
	for _, d := range []string{"easy", "medium", "hard"} {
		if m, ok := r.DifficultyMetrics[d]; ok {
			b.WriteString(fmt.Sprintf("| %s | %d | %.4f | %.4f | %.4f |\n",
				d, m.Count, m.RecallAt10, m.RecallAt20, m.MRR))
		}
	}
	b.WriteString("\n")
}

func renderWorstSamples(b *strings.Builder, r *EvalReport) {
	if len(r.WorstSamples) == 0 {
		return
	}
	b.WriteString("## 5. 异常样本（Worst-10）\n\n> 按 Recall@10 升序排列，便于针对性优化检索策略。\n\n")
	b.WriteString("| SampleID | Query | Topic | Recall@10 | MRR | FirstHitRank |\n|----------|-------|-------|-----------|-----|--------------|\n")
	for _, s := range r.WorstSamples {
		query := s.Query
		if runes := []rune(query); len(runes) > 40 {
			query = string(runes[:40]) + "..."
		}
		query = strings.ReplaceAll(query, "|", "\\|")
		query = strings.ReplaceAll(query, "\n", " ")
		rank := "-"
		if s.FirstHitRank > 0 {
			rank = fmt.Sprintf("%d", s.FirstHitRank)
		}
		b.WriteString(fmt.Sprintf("| %s | %s | %s | %.4f | %.4f | %s |\n",
			s.SampleID, query, s.Topic, s.RecallAt10, s.MRR, rank))
	}
	b.WriteString("\n")
}
