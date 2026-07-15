package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const rerankPrompt = `你是一个文档相关性排序专家。请根据用户查询，对以下【全部】候选文档按相关性从高到低重新排序。

重要规则：必须包含上面所有候选文档的 ID，不要丢弃、不要过滤掉任何文档，只调整它们的先后顺序——把与查询最相关的排在最前面，不太相关的排在后面。

用户查询：%s

候选文档：
%s

请输出重排序后的【完整】文档 ID 列表（JSON 数组格式，从最相关到最不相关，必须包含上面出现的每一个文档 ID），只输出 JSON 数组，不要输出其他内容。
示例输出：["doc_3", "doc_1", "doc_2"]`

// LLMReranker 基于 LLM 的重排序器
type LLMReranker struct {
	model  model.ToolCallingChatModel
	topK   int
	prompt string
}

// NewLLMReranker 创建 LLM 重排序器
func NewLLMReranker(model model.ToolCallingChatModel, topK int) *LLMReranker {
	if topK <= 0 {
		topK = 5
	}
	return &LLMReranker{
		model:  model,
		topK:   topK,
		prompt: rerankPrompt,
	}
}

// Rerank 对文档列表进行重排序
func (reranker *LLMReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= 1 {
		return docs, nil
	}

	// 构造候选文档列表文本
	var docsText string
	docMap := make(map[string]*schema.Document)
	for i, doc := range docs {
		id := doc.ID
		if id == "" {
			id = fmt.Sprintf("doc_%d", i)
		}
		docMap[id] = doc
		// 截取前 200 字符避免 prompt 过长
		content := doc.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		docsText += fmt.Sprintf("[%s] %s\n", id, content)
	}

	prompt := fmt.Sprintf(reranker.prompt, query, docsText)
	messages := []*schema.Message{
		schema.UserMessage(prompt),
	}

	resp, err := reranker.model.Generate(ctx, messages)
	if err != nil {
		// rerank 失败时返回原始顺序
		return docs, nil
	}

	// 解析排序结果
	var rankedIDs []string
	// LLM 返回的是纯数组，不用 extractJSON（它只处理 {}）
	respText := strings.TrimSpace(resp.Content)
	// 提取 [...] 部分
	start := strings.Index(respText, "[")
	end := strings.LastIndex(respText, "]")
	if start >= 0 && end > start {
		respText = respText[start : end+1]
	}
	if err := json.Unmarshal([]byte(respText), &rankedIDs); err != nil {
		// 解析失败返回原始顺序
		return docs, nil
	}

	// 按 LLM 排序重组；LLM 漏掉/误删的按原召回顺序补在末尾，保证不丢任何候选——
	// 这是「全量重排、绝不过滤」的代码层兜底：rerank 只改善排序、不缩小召回集合, 避免「过滤式 rerank」误删相关题拉低 Recall（实测整体 0.7567→0.6533）。
	result := make([]*schema.Document, 0, len(docs))
	seen := make(map[string]bool)
	for _, id := range rankedIDs {
		if doc, ok := docMap[id]; ok && !seen[id] {
			seen[id] = true
			result = append(result, doc)
		}
	}
	for i, doc := range docs {
		id := doc.ID
		if id == "" {
			id = fmt.Sprintf("doc_%d", i)
		}
		if !seen[id] {
			seen[id] = true
			result = append(result, doc)
		}
	}

	// 截断到 topK（先全量重排，再截断）
	if len(result) > reranker.topK {
		result = result[:reranker.topK]
	}
	return result, nil
}

// ScoreBasedReranker 基于分数的简单重排序器
type ScoreBasedReranker struct {
	topK int
}

// NewScoreBasedReranker 创建基于分数的重排序器
func NewScoreBasedReranker(topK int) *ScoreBasedReranker {
	if topK <= 0 {
		topK = 5
	}
	return &ScoreBasedReranker{topK: topK}
}

// Rerank 按 RRF 分数重排序
func (r *ScoreBasedReranker) Rerank(_ context.Context, _ string, docs []*schema.Document) ([]*schema.Document, error) {
	sort.Slice(docs, func(i, j int) bool {
		si, _ := docs[i].MetaData["_rrf_score"].(float64)
		sj, _ := docs[j].MetaData["_rrf_score"].(float64)
		return si > sj
	})

	limit := r.topK
	if limit > len(docs) {
		limit = len(docs)
	}
	return docs[:limit], nil
}

// RerankStrategy 重排策略接口：LLM / cross-encoder / none 三种实现可切换，调用方无感。
type RerankStrategy interface {
	Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error)
}

// rerankEndpoint DashScope 文本排序（cross-encoder）原生 API 地址。
const rerankEndpoint = "https://dashscope.aliyuncs.com/api/v1/services/rerank/text-rerank/text-rerank"

// CrossEncoderReranker 基于 DashScope gte-rerank 专用重排模型（cross-encoder）。
// Go 生态无现成 SDK，这里直接调用 DashScope 原生 rerank HTTP API。
type CrossEncoderReranker struct {
	apiKey     string
	model      string
	topK       int
	endpoint   string // 默认 rerankEndpoint，可在测试中替换为 mock server
	httpClient *http.Client
}

// NewCrossEncoderReranker 创建 cross-encoder 重排器。
// model 为空时用 gte-rerank-v2（gte-rerank 老模型部分账号会返回 403）。
func NewCrossEncoderReranker(apiKey, model string, topK int) *CrossEncoderReranker {
	if topK <= 0 {
		topK = 5
	}
	if model == "" {
		model = "gte-rerank-v2"
	}
	return &CrossEncoderReranker{
		apiKey:     apiKey,
		model:      model,
		topK:       topK,
		endpoint:   rerankEndpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Rerank 调用 DashScope gte-rerank 对候选文档重排；任何失败/错误都降级为原始顺序（截断 topK），不阻断检索。
func (reranker *CrossEncoderReranker) Rerank(ctx context.Context, query string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) <= 1 {
		return docs, nil
	}

	documents := make([]string, len(docs))
	for i, d := range docs {
		documents[i] = d.Content
	}
	reqBody := map[string]any{
		"model": reranker.model,
		"input": map[string]any{"query": query, "documents": documents},
		"parameters": map[string]any{
			"return_documents": false,
			"top_n":            len(docs), // 取全量排序，后面再截断，保证不缩小召回集合
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return reranker.truncate(docs), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reranker.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return reranker.truncate(docs), nil
	}
	req.Header.Set("Authorization", "Bearer "+reranker.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := reranker.httpClient.Do(req)
	if err != nil {
		return reranker.truncate(docs), nil // 网络失败降级
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return reranker.truncate(docs), nil // 403 等降级
	}

	var parsed struct {
		Output struct {
			Results []struct {
				Index int `json:"index"`
			} `json:"results"`
		} `json:"output"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return reranker.truncate(docs), nil
	}

	// 按模型返回顺序映射回文档；漏返的按原顺序补末尾，保证不缩小召回集合（与 LLM 全量重排一致）
	result := make([]*schema.Document, 0, len(docs))
	seen := make(map[int]bool)
	for _, res := range parsed.Output.Results {
		if res.Index >= 0 && res.Index < len(docs) && !seen[res.Index] {
			seen[res.Index] = true
			result = append(result, docs[res.Index])
		}
	}
	for i, d := range docs {
		if !seen[i] {
			seen[i] = true
			result = append(result, d)
		}
	}
	if len(result) > reranker.topK {
		result = result[:reranker.topK]
	}
	return result, nil
}

func (reranker *CrossEncoderReranker) truncate(docs []*schema.Document) []*schema.Document {
	if len(docs) > reranker.topK {
		return docs[:reranker.topK]
	}
	return docs
}

// NoneReranker 不重排
type NoneReranker struct {
	topK int
}

// NewNoneReranker 创建不重排策略
func NewNoneReranker(topK int) *NoneReranker {
	if topK <= 0 {
		topK = 5
	}
	return &NoneReranker{topK: topK}
}

// Rerank 直接返回原始顺序
func (reranker *NoneReranker) Rerank(_ context.Context, _ string, docs []*schema.Document) ([]*schema.Document, error) {
	if len(docs) > reranker.topK {
		return docs[:reranker.topK], nil
	}
	return docs, nil
}
