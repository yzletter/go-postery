package index_service

import (
	"errors"
	"strings"
	"sync/atomic"

	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/backend/micro/search/internal/forward_index"
	reverse_index2 "github.com/yzletter/go-postery/backend/micro/search/internal/reverse_index"
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
)

// Indexer 搜索引擎索引
type Indexer struct {
	// 正倒排索引结合
	forwardIndex forward_index.ForwardIndex
	reverseIndex reverse_index2.ReverseIndex
	maxIndexID   uint64
}

// Init 初始化索引
func (indexer *Indexer) Init(DocNumEstimate int, DataDir string) error {
	db, err := forward_index.NewForwardIndex(DataDir)
	if err != nil {
		return err
	}
	indexer.forwardIndex = db
	indexer.reverseIndex = reverse_index2.NewSkipListReverseIndex(DocNumEstimate)
	return nil
}

// Close 关闭索引
func (indexer *Indexer) Close() error {
	return indexer.forwardIndex.Close()
}

func (indexer *Indexer) AddDoc(doc *model2.Document) (int, error) {
	// 参数校验
	if doc == nil {
		return 0, errors.New("empty search document")
	}
	docID := strings.TrimSpace(doc.DocID)
	if len(docID) == 0 {
		return 0, errors.New("empty search document id")
	}
	doc.DocID = docID

	// 获取索引 ID
	doc.IndexID = atomic.AddUint64(&indexer.maxIndexID, 1)

	// 序列化后写入正排
	key := []byte(docID)
	value, err := sonic.Marshal(doc)
	if err != nil {
		return 0, err
	}
	if err = indexer.forwardIndex.Set(key, value); err != nil {
		return 0, err
	}

	// 写入倒排
	indexer.reverseIndex.Add(doc)
	return 1, nil
}

func (indexer *Indexer) UpdateDoc(doc *model2.Document) (int, error) {
	docID := strings.TrimSpace(doc.DocID)
	if len(docID) == 0 {
		return 0, nil
	}

	indexer.DeleteDoc(docID)
	return indexer.AddDoc(doc)
}

func (indexer *Indexer) DeleteDoc(docID string) int {
	docBytes, err := indexer.forwardIndex.Get([]byte(docID))
	if err != nil {
		return 0
	} else if len(docBytes) == 0 {
		return 0
	}

	var document model2.Document
	if err = sonic.Unmarshal(docBytes, &document); err != nil {
		slog.Warn("invalid search document, skip delete", "doc_id", docID, "error", err)
		return 0
	}

	// 删倒排
	for _, keyword := range document.Keywords {
		indexer.reverseIndex.Del(document.IndexID, keyword)
	}

	// 删正排
	if err = indexer.forwardIndex.Delete([]byte(docID)); err != nil {
		slog.Error("delete forward index failed", "doc_id", docID, "error", err)
		return 0
	}

	return 1
}

func (indexer *Indexer) Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []*model2.Document {
	// 搜倒排
	docIDs := indexer.reverseIndex.Search(query, onFlag, offFlag, orFlags)
	if len(docIDs) == 0 {
		return nil
	}
	// 构造批量 Keys
	keys := make([][]byte, 0, len(docIDs))
	for _, docID := range docIDs {
		keys = append(keys, []byte(docID))
	}

	// 找正排
	docs, err := indexer.forwardIndex.BatchGet(keys)
	if err != nil {
		slog.Error("read forward index failed", "doc_count", len(keys), "error", err)
		return nil
	}

	res := make([]*model2.Document, 0, len(docIDs))
	for i, docBytes := range docs {
		var document model2.Document
		err = sonic.Unmarshal(docBytes, &document)
		if err != nil {
			slog.Warn("invalid search document, skip", "doc_id", string(keys[i]), "error", err)
			continue
		}
		res = append(res, &document)
	}

	return res
}

// Count 索引里有多少 Document
func (indexer *Indexer) Count() int {
	cnt := 0
	indexer.forwardIndex.IterKey(func(k []byte) error {
		cnt++
		return nil
	})
	return cnt
}

// LoadFromIndexFile 系统重启时，直接从正排里加载数据
func (indexer *Indexer) LoadFromIndexFile() int {
	loadedCount := 0
	indexer.forwardIndex.IterDB(func(k, v []byte) error {
		var document model2.Document
		if err := sonic.Unmarshal(v, &document); err != nil {
			slog.Warn("invalid search index file document, skip", "doc_id", string(k), "error", err)
			return nil
		}
		indexer.reverseIndex.Add(&document)
		loadedCount++
		return nil
	})

	return loadedCount
}
