package forward_index

import (
	"errors"
	"sync/atomic"

	bolt "go.etcd.io/bbolt"
)

// BoltForwardIndex 用 Bolt 实现 ForwardIndex
type BoltForwardIndex struct {
	db     *bolt.DB
	path   string
	bucket []byte
}

func (index *BoltForwardIndex) AddFilePath(filepath string) *BoltForwardIndex {
	index.path = filepath
	return index
}

func (index *BoltForwardIndex) AddBucket(bucket string) *BoltForwardIndex {
	index.bucket = []byte(bucket)
	return index
}

func (index *BoltForwardIndex) Open() error {
	// 打开 BoltDB
	db, err := bolt.Open(index.GetDbPath(), 0o600, bolt.DefaultOptions)
	if err != nil {
		return err
	}

	// 初始化存储桶
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(index.bucket)
		return err
	}); err != nil {
		_ = db.Close()
		return err
	}

	index.db = db
	return nil
}

func (index *BoltForwardIndex) GetDbPath() string {
	return index.path
}

func (index *BoltForwardIndex) Set(k, v []byte) error {
	return index.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(index.bucket).Put(k, v)
		return err
	})
}

func (index *BoltForwardIndex) BatchSet(keys, values [][]byte) error {
	if len(keys) != len(values) {
		return errors.New("key value not the same length")
	}

	index.db.Batch(func(tx *bolt.Tx) error {
		for i, key := range keys {
			value := values[i]
			_ = tx.Bucket(index.bucket).Put(key, value)
		}
		return nil
	})
	return nil
}

func (index *BoltForwardIndex) Get(k []byte) ([]byte, error) {
	var value []byte
	index.db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket(index.bucket).Get(k)
		return nil
	})

	if len(value) == 0 {
		return nil, errors.New("no data")
	}

	return value, nil
}

func (index *BoltForwardIndex) BatchGet(keys [][]byte) ([][]byte, error) {
	values := make([][]byte, len(keys))
	index.db.Batch(func(tx *bolt.Tx) error {
		for i, key := range keys {
			value := tx.Bucket(index.bucket).Get(key)
			values[i] = value
		}
		return nil
	})

	return values, nil
}

func (index *BoltForwardIndex) Delete(k []byte) error {
	return index.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(index.bucket).Delete(k)
		return err
	})
}

func (index *BoltForwardIndex) BatchDelete(keys [][]byte) error {
	index.db.Batch(func(tx *bolt.Tx) error {
		for _, key := range keys {
			_ = tx.Bucket(index.bucket).Delete(key)
		}
		return nil
	})

	return nil
}

func (index *BoltForwardIndex) Has(k []byte) bool {
	var value []byte
	err := index.db.View(func(tx *bolt.Tx) error {
		value = tx.Bucket(index.bucket).Get(k)
		return nil
	})

	// 没有读到数据
	if err != nil || string(value) == "" {
		return false
	}

	return true
}

func (index *BoltForwardIndex) IterDB(fn func(k []byte, v []byte) error) int64 {
	var total int64
	index.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(index.bucket).Cursor() // 获取游标进行遍历
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			if err := fn(k, v); err != nil {
				return err
			}
			atomic.AddInt64(&total, 1)
		}
		return nil
	})

	return total
}

func (index *BoltForwardIndex) IterKey(fn func(k []byte) error) int64 {
	var total int64
	index.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(index.bucket).Cursor() // 获取游标进行遍历
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			if err := fn(k); err != nil {
				return err
			}
			atomic.AddInt64(&total, 1)
		}
		return nil
	})

	return total
}

func (index *BoltForwardIndex) Close() error {
	return index.db.Close()
}

func (index *BoltForwardIndex) WALName() string {
	return index.db.Path()
}
