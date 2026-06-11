package reverse_index

import (
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
)

// ReverseIndex 倒排索引接口
type ReverseIndex interface {
	Add(document *model2.Document)
	Del(IndexID uint64, keyword *model2.Keyword)
	Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []string
}
