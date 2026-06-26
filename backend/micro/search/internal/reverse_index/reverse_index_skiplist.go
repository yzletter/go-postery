package reverse_index

import (
	"runtime"
	"sync"

	"github.com/huandu/skiplist"
	farmhash "github.com/leemcloughlin/gofarmhash"
	model2 "github.com/yzletter/go-postery/backend/micro/search/model"
	"github.com/yzletter/go-postery/backend/utils"
)

// SkipListReverseIndex 用跳表实现的倒排索引
type SkipListReverseIndex struct {
	table *utils.ConcurrentMap // 手写并发安全 Map, 每个 Key 对应一条跳表
	locks []sync.RWMutex       // 当操作同一条倒排链时要加锁
}

// SkipListNodeKey 定义跳表节点的 Key 即 Document 的 IndexID
type SkipListNodeKey uint64

// SkipListNodeValue 定义跳表节点的 Value
type SkipListNodeValue struct {
	DocID      string // 即 Document.ID ——> 业务 ID
	DocFeature uint64 // 表示文档特征的 Bits
}

// NewSkipListReverseIndex 构造函数传入预估 Document 数量
func NewSkipListReverseIndex(DocumentNumEstimate int) *SkipListReverseIndex {
	// 按 CPU 核数拆分小 Map, 降低并发写冲突
	mp := utils.NewConcurrentMap(runtime.NumCPU()*500, DocumentNumEstimate)
	locks := make([]sync.RWMutex, 1000)
	return &SkipListReverseIndex{
		table: mp,
		locks: locks,
	}
}

// Add 添加倒排索引
func (index *SkipListReverseIndex) Add(document *model2.Document) {
	// 获取文档关键词列表
	keywords := document.Keywords

	// 给每个关键词的倒排链加上当前文档
	for _, keyword := range keywords {
		key := keyword.ToString()

		// 同一个关键词固定落到同一把锁
		lock := index.getLock(key)
		lock.Lock()

		// 写入跳表节点 Value
		nodeValue := SkipListNodeValue{DocID: document.DocID, DocFeature: document.BitsFeature}

		if value, exist := index.table.Get(key); !exist { // 不存在当前 Keyword 对应链表
			// 新建一条跳表
			list := skiplist.New(skiplist.Uint64)
			list.Set(document.IndexID, nodeValue)

			// 把新建的跳表加入 Map 中
			index.table.Set(key, list)
		} else { // 存在链表
			// 断言成跳表
			list := value.(*skiplist.SkipList)
			list.Set(document.IndexID, nodeValue)
		}

		lock.Unlock()
	}
}

// Del 删除 Keyword 链上的文档
func (index *SkipListReverseIndex) Del(IndexID uint64, keyword *model2.Keyword) {
	key := keyword.ToString()
	lock := index.getLock(key)

	// 加锁
	lock.Lock()
	defer lock.Unlock()

	// 获取对应倒排链
	if value, exist := index.table.Get(key); exist {
		// 断言成跳表
		list := value.(*skiplist.SkipList)
		list.Remove(IndexID)
	}
}

// Search 搜索, 返回文档的 DocID 即业务 ID
func (index *SkipListReverseIndex) Search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) []string {
	result := index.search(query, onFlag, offFlag, orFlags)
	if result == nil {
		return nil
	}
	arr := make([]string, 0, result.Len())
	node := result.Front()
	for node != nil {
		skv, _ := node.Value.(SkipListNodeValue)
		arr = append(arr, skv.DocID)
		node = node.Next()
	}
	return arr
}

// 根据 keyword 哈希分配一把锁，操作相同 keyword 不会产生冲突
func (index *SkipListReverseIndex) getLock(keyword string) *sync.RWMutex {
	idx := int(farmhash.Hash32WithSeed([]byte(keyword), 0))
	return &index.locks[idx%len(index.locks)]
}

// 根据 BitsFeature 和条件进行过滤
func (index *SkipListReverseIndex) filterByDocFeature(docFeature uint64, onFlag uint64, offFlag uint64, orFlags []uint64) bool {
	if docFeature&onFlag != onFlag {
		return false
	}

	if docFeature&offFlag != 0 {
		return false
	}

	for _, orFlag := range orFlags {
		if orFlag > 0 && docFeature&orFlag <= 0 {
			return false
		}
	}

	return true
}

func (index *SkipListReverseIndex) search(query *model2.TermQuery, onFlag uint64, offFlag uint64, orFlags []uint64) *skiplist.SkipList {
	if query.Keyword != nil { // 当前 Query 就是一个关键词
		keyword := query.Keyword.ToString()
		// 抢锁
		lock := index.getLock(keyword)
		lock.RLock()
		defer lock.RUnlock()

		// 获取倒排链
		if value, exist := index.table.Get(keyword); exist {
			// 倒排链存在
			list := value.(*skiplist.SkipList)
			res := skiplist.New(skiplist.Uint64) // 答案

			// 遍历倒排链并按 BitsFeature 过筛
			node := list.Front()
			for node != nil {
				indexID := node.Key().(uint64)
				slnv := node.Value.(SkipListNodeValue)
				if indexID > 0 && index.filterByDocFeature(slnv.DocFeature, onFlag, offFlag, orFlags) {
					// 满足条件, 记录答案
					res.Set(indexID, slnv)
				}
				node = node.Next()
			}
			return res
		}
	} else if len(query.Must) > 0 {
		// Must 条件取交集
		lists := make([]*skiplist.SkipList, 0, len(query.Must))
		for _, q := range query.Must {
			list := index.search(q, onFlag, offFlag, orFlags)
			lists = append(lists, list)
		}
		return intersectionOfSkipList(lists...)
	} else if len(query.Should) > 0 {
		// Should 条件取并集
		lists := make([]*skiplist.SkipList, 0, len(query.Should))
		for _, q := range query.Should {
			list := index.search(q, onFlag, offFlag, orFlags)
			lists = append(lists, list)
		}
		return unionSetOfSkipList(lists...)
	}
	return nil
}

// intersectionOfSkipList 多跳表求交集
func intersectionOfSkipList(lists ...*skiplist.SkipList) *skiplist.SkipList {
	if len(lists) == 0 {
		return nil
	} else if len(lists) == 1 {
		return lists[0]
	}

	nodes := make([]*skiplist.Element, len(lists))
	for i, list := range lists {
		// 有空跳表不会有交集
		if list == nil || list.Len() == 0 {
			return nil
		}

		// 初始时都指向第一个节点
		nodes[i] = list.Front()
	}

	res := skiplist.New(skiplist.Uint64)
	maxm := make(map[int]struct{}, len(nodes)) // 记录当前等于 maxKey 的节点个数
	var maxKey uint64 = 0

	for {
		// 遍历所有节点
		for idx, node := range nodes {
			nowKey := node.Key().(uint64)
			if nowKey > maxKey {
				maxKey = nowKey
				maxm = map[int]struct{}{idx: {}} // 清空集合并标记当前下标
			} else if nowKey == maxKey {
				maxm[idx] = struct{}{}
			}
		}

		if len(maxm) == len(nodes) { // 当前所有节点都等于 maxKey
			// 记录答案
			res.Set(nodes[0].Key(), nodes[0].Value)

			// 所有指针后移一步
			for idx := range nodes { // 不能用 node = node.Next() 是值拷贝
				nodes[idx] = nodes[idx].Next()
				if nodes[idx] == nil {
					return res
				}
			}
		} else {
			for idx := range nodes { // 不能用 node = node.Next() 是值拷贝
				for nodes[idx].Key().(uint64) < maxKey {
					nodes[idx] = nodes[idx].Next()
					if nodes[idx] == nil {
						return res
					}
				}
				nodes[idx] = nodes[idx].Next()
				if nodes[idx] == nil {
					return res
				}
			}
		}
	}
}

// unionSetOfSkipList 多跳表求并集
func unionSetOfSkipList(lists ...*skiplist.SkipList) *skiplist.SkipList {
	if len(lists) == 0 {
		return nil
	} else if len(lists) == 1 {
		return lists[0]
	}

	res := skiplist.New(skiplist.Uint64)
	for _, list := range lists {
		if list == nil {
			continue
		}
		node := list.Front()
		for node != nil {
			res.Set(node.Key(), node.Value) // github.com/huandu/skiplist 会处理相同 key
			node = node.Next()
		}
	}
	return res
}
