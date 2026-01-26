package reverse_index

import "github.com/yzletter/go-postery/infra/search/model"

// ReverseIndex 倒排索引接口
type ReverseIndex interface {
	Add(document *model.Document)
	Del(IndexID uint64, keyword *model.Keyword)
	Search(query *model.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []string
}
