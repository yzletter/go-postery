package rag

import (
	"context"
	"log/slog"
	"sort"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
)

const RRFConstant = 60

// MultiRetriever 多路召回器
type MultiRetriever struct {
	retrievers []retriever.Retriever // 多路召回
	topK       int
	k          int // 参数
}

func NewMultiRetriever(retrievers []retriever.Retriever, topK int) *MultiRetriever {
	return &MultiRetriever{
		retrievers: retrievers,
		topK:       topK,
		k:          RRFConstant,
	}
}

func (m *MultiRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	res := make([][]*schema.Document, len(m.retrievers))
	// 多路找回后汇总
	for idx, r := range m.retrievers {
		documents, err := r.Retrieve(ctx, query, opts...)
		if err != nil {
			slog.Error("retrieve failed", "error", err)
			continue
		}
		res[idx] = append(res[idx], documents...)
	}

	// 融合
	return m.rrfFusion(res), nil
}

// score(d) = Σ 1 / (k + rank_i(d)) =========> 文章 d 的分数 = (参数 + rank_i(d) 即文章在某路召回中的排名) 倒数求和
func (m *MultiRetriever) rrfFusion(all [][]*schema.Document) []*schema.Document {
	scores := make(map[string]float64)          // 文章分数
	docMap := make(map[string]*schema.Document) // 最后最近一次的文档进行保留

	for _, docs := range all {
		for rank, doc := range docs {
			id := docID(doc)
			scores[id] += 1 / float64(m.k+rank+1)
			docMap[id] = doc
		}
	}

	// 结构体排序
	type ScoredDoc struct {
		id    string
		score float64
	}

	scoredDocs := make([]ScoredDoc, 0, len(scores))
	for id, score := range scores {
		if score > 0 {
			scoredDocs = append(scoredDocs, ScoredDoc{id, score})
		}
	}

	// 降序排列
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].score > scoredDocs[j].score
	})

	// 取 topK
	limit := m.topK
	if limit > len(scoredDocs) {
		limit = len(scoredDocs)
	}

	results := make([]*schema.Document, limit)
	for i := 0; i < limit; i++ {
		doc := docMap[scoredDocs[i].id]
		docCopy := *doc
		if docCopy.MetaData == nil {
			docCopy.MetaData = make(map[string]any)
		}
		docCopy.MetaData["_rrf_score"] = scoredDocs[i].score
		results[i] = &docCopy
	}

	return results
}

// docID 获取文档唯一标识
func docID(doc *schema.Document) string {
	if doc.ID != "" {
		return doc.ID
	}
	// 没有 ID 时用内容前 100 个字符作为标识
	if len(doc.Content) > 100 {
		return doc.Content[:100]
	}
	return doc.Content
}
