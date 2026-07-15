package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/yzletter/go-postery/backend/ports"
)

type BM25Retriever struct {
	tokenizer ports.Tokenizer    // 分词器
	documents []*schema.Document // 所储存的文档
	indexer   map[string][]docTF // 倒排索引 (词 ——> (docIndex, 词频))
	docLen    []int              // 每篇文档的词数
	avgDL     float64            // 平均文档词数
	topK      int                // 召回文章数
	k1        float64            // BM25 参数，控制词频饱和度
	b         float64            // BM25 参数，控制文档长度归一化
}

type docTF struct {
	docIdx int // 下标
	tf     int // 词频
}

// NewBM25Retriever 创建 BM25 检索器
func NewBM25Retriever(tokenizer ports.Tokenizer, topK int) *BM25Retriever {
	if topK <= 0 {
		topK = 10
	}
	return &BM25Retriever{
		tokenizer: tokenizer,
		indexer:   make(map[string][]docTF),
		topK:      topK,
		k1:        1.5, // 业界优秀参数
		b:         0.75,
	}
}

// GetDocuments 获取当前索引中的所有文档
func (retriever *BM25Retriever) GetDocuments() []*schema.Document {
	return retriever.documents
}

// IndexDocuments 建立索引
func (retriever *BM25Retriever) IndexDocuments(docs []*schema.Document) {
	// 当前索引所存储的文档
	retriever.documents = docs // 所有文档
	retriever.indexer = make(map[string][]docTF)
	retriever.docLen = make([]int, len(docs)) // 每篇文档词数

	// 遍历文档
	totalLen := 0 // 文章总词数
	for idx, doc := range docs {
		// 分词
		tokens := filterBM25Tokens(retriever.tokenizer.CutSearch(doc.Content))
		// 当前文档词数
		retriever.docLen[idx] = len(tokens)
		totalLen += len(tokens)
		// 统计词频
		cnt := make(map[string]int)
		for _, token := range tokens {
			cnt[token]++
		}
		// 放入索引
		for k, v := range cnt {
			retriever.indexer[k] = append(retriever.indexer[k], docTF{docIdx: idx, tf: v})
		}
	}

	// 统计平均词数
	if len(docs) > 0 {
		retriever.avgDL = float64(totalLen) / float64(len(docs))
	}
}

// Retrieve 召回, 实现 eino retriever.Retriever 接口
func (retriever *BM25Retriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	if len(retriever.documents) == 0 {
		return nil, nil
	}

	// 分词
	tokens := filterBM25Tokens(retriever.tokenizer.Cut(query))
	n := len(retriever.documents) // 总文章数
	scores := make([]float64, n)  // 每篇文章的分数

	// 遍历分词
	for _, token := range tokens {
		docs, exists := retriever.indexer[token] // 该词的索引
		if !exists {
			continue
		}
		df := float64(len(docs)) // df 有该 token 的文章数
		// IDF = log((N - df + 0.5) / (df + 0.5) + 1)
		IDF := math.Log((float64(n)-df+0.5)/(df+0.5) + 1)

		// 遍历每篇文章
		for _, doc := range docs {
			tf := float64(doc.tf)                       // 当前文章该词词频
			dl := float64(retriever.docLen[doc.docIdx]) // 当前文章总词数
			// BM25 score = IDF * (tf * (k1+1)) / (tf + k1 * (1 - b + b * dl/avgDL))
			bm25score := IDF * (tf * (retriever.k1 + 1)) / (tf + retriever.k1*(1-retriever.b+retriever.b*dl/retriever.avgDL))
			scores[doc.docIdx] += bm25score // 累加该词的贡献
		}
	}

	type ScoredDoc struct {
		idx   int
		score float64
	}

	scoredDocs := make([]ScoredDoc, 0, len(scores))
	for idx, score := range scores {
		if score > 0 {
			scoredDocs = append(scoredDocs, ScoredDoc{idx, score})
		}
	}

	// 降序排列
	sort.Slice(scoredDocs, func(i, j int) bool {
		return scoredDocs[i].score > scoredDocs[j].score
	})

	// 取 topK
	limit := retriever.topK
	if limit > len(scoredDocs) {
		limit = len(scoredDocs)
	}

	// 返回结果
	res := make([]*schema.Document, limit)
	for i := 0; i < limit; i++ {
		doc := retriever.documents[scoredDocs[i].idx]
		// 将 BM25 分数存入 metadata
		docCopy := *doc
		// 不存在 metadata 就新建
		if docCopy.MetaData == nil {
			docCopy.MetaData = make(map[string]any)
		}
		docCopy.MetaData["_bm25_score"] = scoredDocs[i].score
		res[i] = &docCopy
	}

	return res, nil
}

func filterBM25Tokens(tokens []string) []string {
	filtered := make([]string, 0, len(tokens))
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if !hasLetterOrNumber(token) {
			continue
		}
		filtered = append(filtered, token)
	}
	return filtered
}

func hasLetterOrNumber(token string) bool {
	for _, r := range token {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

// BM25Doc BM25 文档加载用的中间结构
type BM25Doc struct {
	ID        int64  `json:"id"` // 题目 ID
	Content   string `json:"content"`
	Reference string `json:"reference"`
}

// LoadBM25DocsFromFile 从 JSON 文件加载文档
func LoadBM25DocsFromFile(filePath string) ([]*BM25Doc, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("bm25: read %s: %w", filePath, err)
	}
	var docs []*BM25Doc
	if err := json.Unmarshal(data, &docs); err != nil {
		return nil, fmt.Errorf("bm25: parse %s: %w", filePath, err)
	}
	return docs, nil
}

// AppendParsedQuestions 将解析后的题目追加到 BM25 索引
func (retriever *BM25Retriever) AppendParsedQuestions(questions []BM25Doc) {
	newDocs := make([]*schema.Document, len(questions))
	for i, q := range questions {
		newDocs[i] = &schema.Document{
			// schema.Document.ID 是 string, 业务 ID 在边界处格式化
			ID:      fmt.Sprintf("%d", q.ID),
			Content: q.Content + "\n参考答案：" + q.Reference,
		}
	}
	// 追加到现有文档后重建索引
	retriever.documents = append(retriever.documents, newDocs...)
	allDocs := retriever.documents
	retriever.IndexDocuments(allDocs)
}

// IndexBM25Docs 从 BM25Doc 构建索引
func (retriever *BM25Retriever) IndexBM25Docs(docs []*BM25Doc) {
	schemaDocs := make([]*schema.Document, len(docs))
	for i, d := range docs {
		schemaDocs[i] = &schema.Document{
			// schema.Document.ID 是 string, 业务 ID 在边界处格式化
			ID:      fmt.Sprintf("%d", d.ID),
			Content: d.Content + "\n参考答案：" + d.Reference,
		}
	}
	retriever.IndexDocuments(schemaDocs)
}

// BM25Manager 按用户管理 BM25 索引缓存
type BM25Manager struct {
	tokenizer  ports.Tokenizer
	mu         sync.RWMutex
	retrievers map[int64]*BM25Retriever // userID -> retriever
	topK       int
}

// NewBM25Manager 创建 BM25 管理器
func NewBM25Manager(tokenizer ports.Tokenizer, topK int) *BM25Manager {
	if topK <= 0 {
		topK = 10
	}
	return &BM25Manager{
		tokenizer:  tokenizer,
		retrievers: make(map[int64]*BM25Retriever),
		topK:       topK,
	}
}

// ReplaceDocuments 覆盖指定用户的题库索引
func (cache *BM25Manager) ReplaceDocuments(userID int64, docs []*schema.Document) {
	// 建立索引
	r := NewBM25Retriever(cache.tokenizer, cache.topK)
	r.IndexDocuments(docs)
	// 更新缓存
	cache.mu.Lock()
	cache.retrievers[userID] = r
	cache.mu.Unlock()
}

// AppendDocuments 追加文档到指定用户的索引
func (cache *BM25Manager) AppendDocuments(userID int64, docs []*schema.Document) {
	// 当前用户是否有索引
	cache.mu.Lock()
	r, ok := cache.retrievers[userID]
	if !ok {
		r = NewBM25Retriever(cache.tokenizer, cache.topK)
		cache.retrievers[userID] = r
	}
	cache.mu.Unlock()
	r.documents = append(r.documents, docs...) // 追加文档
	r.IndexDocuments(r.documents)              // 对所有文档重建索引
}

// Retrieve 检索指定用户的题库
func (cache *BM25Manager) Retrieve(ctx context.Context, userID int64, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	cache.mu.RLock()
	r, ok := cache.retrievers[userID]
	cache.mu.RUnlock()
	if !ok {
		return nil, nil
	}

	return r.Retrieve(ctx, query, opts...)
}

// DeleteByUserID 删除指定用户的索引
func (cache *BM25Manager) DeleteByUserID(userID int64) {
	cache.mu.Lock()
	delete(cache.retrievers, userID)
	cache.mu.Unlock()
}
