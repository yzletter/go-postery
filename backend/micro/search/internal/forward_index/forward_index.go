package forward_index

import (
	"os"
	"strings"
)

type ForwardIndex interface {
	// Open 初始化 DB
	//
	// Return:
	//	- error: 可能返回的错误
	Open() error

	// GetDbPath 获取存储数据的目录
	//
	// Return:
	//	- string: 数据目录
	GetDbPath() string

	// Set 写入 <key, value>
	//
	// Parameter:
	//	- k: key
	//	- v: value
	//
	// Return:
	//	- error: 可能返回的错误
	Set(k, v []byte) error

	// BatchSet 批量写入 <key, value>
	//
	// Parameter:
	//	- keys: key 列表
	//	- values: value 列表
	//
	// Return:
	//	- error: 可能返回的错误
	BatchSet(keys, values [][]byte) error

	// Get 读取 key 对应的 value
	//
	// Parameter:
	//	- k: key
	//
	// Return:
	//	- []byte: value
	//	- error: 可能返回的错误
	Get(k []byte) ([]byte, error)

	// BatchGet 批量读取
	//
	// Parameter:
	//	- keys: key 列表
	//
	// Return:
	//	- [][]byte: value 列表
	//	- error: 可能返回的错误
	BatchGet(keys [][]byte) ([][]byte, error)

	// Delete 删除
	//
	// Parameter:
	//	- k: key
	//
	// Return:
	//	- error: 可能返回的错误
	Delete(k []byte) error

	// BatchDelete 批量删除
	//
	// Parameter:
	//	- keys: key 列表
	//
	// Return:
	//	- error: 可能返回的错误
	BatchDelete(keys [][]byte) error

	// Has 判断 key 是否存在
	//
	// Parameter:
	//	- k: key
	//
	// Return:
	//	- bool: key 是否存在
	Has(k []byte) bool

	// IterDB 遍历数据库
	//
	// Parameter:
	//	- fn: 遍历回调
	//
	// Return:
	//	- int64: 数据条数
	IterDB(fn func(k, v []byte) error) int64

	// IterKey 遍历所有 key
	//
	// Parameter:
	//	- fn: 遍历回调
	//
	// Return:
	//	- int64: key 数量
	IterKey(fn func(k []byte) error) int64

	// Close 把内存中的数据 flush 到磁盘，同时释放文件锁
	//
	// Return:
	//	- error: 可能返回的错误
	Close() error
}

// NewForwardIndex 根据文件路径创建正排索引
func NewForwardIndex(filepath string) (ForwardIndex, error) {
	// 文件路径校验
	paths := strings.Split(filepath, "/")
	parentPath := strings.Join(paths[0:len(paths)-1], "/") // 父路径

	info, err := os.Stat(parentPath)
	if os.IsNotExist(err) { // 父路径不存在则创建
		_ = os.MkdirAll(parentPath, os.ModePerm)
	} else {
		// 父路径存在
		if info.Mode().IsRegular() { // 父路径是普通文件时先删除
			_ = os.Remove(parentPath)
		}
	}

	index := new(BoltForwardIndex)
	index.AddFilePath(filepath).AddBucket("go-postery")

	err = index.Open()
	if err != nil {
		return nil, err
	}
	return index, nil
}
