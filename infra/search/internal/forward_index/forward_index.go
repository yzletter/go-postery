package forward_index

import (
	"os"
	"strings"
)

type ForwardIndex interface {
	Open() error                              // 初始化 DB
	GetDbPath() string                        // 获取存储数据的目录
	Set(k, v []byte) error                    // 写入 <key, value>。Document 的 DocID 作为 key，Document 作为 value。
	BatchSet(keys, values [][]byte) error     // 批量写入 <key, value>
	Get(k []byte) ([]byte, error)             // 读取 key 对应的 value
	BatchGet(keys [][]byte) ([][]byte, error) // 批量读取，注意不保证顺序
	Delete(k []byte) error                    // 删除
	BatchDelete(keys [][]byte) error          // 批量删除
	Has(k []byte) bool                        // 判断 key 是否存在
	IterDB(fn func(k, v []byte) error) int64  // 遍历数据库，返回数据的条数
	IterKey(fn func(k []byte) error) int64    // 遍历所有key，返回数据的条数
	Close() error                             // 把内存中的数据flush到磁盘，同时释放文件锁
}

// NewForwardIndex 工厂模式
func NewForwardIndex(filepath string) (ForwardIndex, error) {
	// 文件路径校验
	paths := strings.Split(filepath, "/")
	parentPath := strings.Join(paths[0:len(paths)-1], "/") // 父路径

	info, err := os.Stat(parentPath)
	if os.IsNotExist(err) { //如果父路径不存在则创建
		_ = os.MkdirAll(parentPath, os.ModePerm) //数字前的0或0o都表示八进制
	} else {
		// 父路径存在
		if info.Mode().IsRegular() { //如果父路径是个普通文件，则把它删掉
			_ = os.Remove(parentPath)
		}
	}

	index := new(BoltForwardIndex)
	index.AddFilePath(filepath).AddBucket("go-searchery")

	err = index.Open()
	if err != nil {
		return nil, err
	}
	return index, nil
}
