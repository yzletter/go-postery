package reverse_index

import (
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
)

// ReverseIndex 倒排索引接口
type ReverseIndex interface {
	// Add 添加文档到倒排索引
	//
	// Parameter:
	//	- document: 文档
	Add(document *model2.Document)

	// Del 从倒排索引删除关键词
	//
	// Parameter:
	//	- IndexID: 索引 ID
	//	- keyword: 关键词
	Del(IndexID uint64, keyword *model2.Keyword)

	// Search 搜索倒排索引
	//
	// Parameter:
	//	- query: 查询条件
	//	- onFlag: 必须开启的标志
	//	- offFlag: 必须关闭的标志
	//	- orFlags: 任一匹配标志列表
	//
	// Return:
	//	- []string: 命中的 DocID 列表
	Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []string
}
