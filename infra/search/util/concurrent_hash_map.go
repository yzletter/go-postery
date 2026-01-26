package util

import (
	"math/rand"
	"sync"

	"github.com/leemcloughlin/gofarmhash"
	"golang.org/x/exp/maps"
)

// ConcurrentMap 并发安全 Map
type ConcurrentMap struct {
	mps   []map[string]any // 许多小 map
	count int              // 小 map 数
	locks []sync.RWMutex   // 每个小 map 对应的锁
	seed  uint32           // 用于哈希的种子
}

// NewConcurrentMap 构造函数, 传入小map数和预估容量
func NewConcurrentMap(count, cap int) *ConcurrentMap {
	mps := make([]map[string]any, count)
	for i := 0; i < count; i++ {
		mps[i] = make(map[string]any, cap/count) // 每个小 mp 的大小
	}
	locks := make([]sync.RWMutex, cap)
	seed := rand.Uint32()

	return &ConcurrentMap{
		mps:   mps,
		count: count,
		locks: locks,
		seed:  seed,
	}
}

// Set 将 < Key, Value > 放入 Map
func (mp *ConcurrentMap) Set(key string, value any) {
	// 获取下标
	idx := mp.getIndex(key)

	// 上写锁
	mp.locks[idx].Lock()
	// defer 解写锁
	defer mp.locks[idx].Unlock()

	// 存入小 map
	mp.mps[idx][key] = value
}

// Get 从 Map 中获取 Key 对应的 Value
func (mp *ConcurrentMap) Get(key string) (any, bool) {
	// 获取下标
	idx := mp.getIndex(key)

	// 上读锁
	mp.locks[idx].RLock()
	// defer 解读锁
	defer mp.locks[idx].RUnlock()

	value, exists := mp.mps[idx][key]
	return value, exists
}

// 根据 Key 获取对应小 map 下标
func (mp *ConcurrentMap) getIndex(key string) int {
	return int(farmhash.Hash32WithSeed([]byte(key), mp.seed)) % mp.count
}

// NewConcurrentMapIterator 构造迭代器
func (mp *ConcurrentMap) NewConcurrentMapIterator() *ConcurrentMapIterator {
	// 初始化
	keys := make([][]string, 0, len(mp.mps))

	// 遍历每个小 Map
	for _, internal := range mp.mps {
		res := maps.Keys(internal) // 获取当前 Map 所有的 Keys
		keys = append(keys, res)
	}

	return &ConcurrentMapIterator{
		concurrentMap: mp,
		keys:          keys,
		rowIndex:      0,
		colIndex:      0,
	}
}

// MapIterator 迭代器模式
type MapIterator interface {
	Next() any
}

type ConcurrentMapIterator struct {
	concurrentMap *ConcurrentMap
	keys          [][]string
	rowIndex      int
	colIndex      int
}

type MapEntry struct {
	Key   string
	Value any
}

func (iter *ConcurrentMapIterator) Next() *MapEntry {
	// 遍历完
	if iter.rowIndex >= len(iter.keys) {
		return nil
	}

	// 当前行对应的小 Map 为空
	row := iter.keys[iter.rowIndex]
	if len(row) == 0 {
		iter.rowIndex++
		return iter.Next()
	}

	// 当前遍历到的 Key
	key := row[iter.colIndex]
	value, _ := iter.concurrentMap.Get(key)

	// 调整下标
	if iter.colIndex+1 >= len(iter.keys[iter.rowIndex]) {
		iter.rowIndex++
		iter.colIndex = 0
	} else {
		iter.colIndex++
	}

	return &MapEntry{
		Key:   key,
		Value: value,
	}
}
