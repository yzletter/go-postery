package etcd

import (
	"log/slog"
	"sync"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

var (
	client *etcdv3.Client
	once   sync.Once
)

// Init 单例模式初始化 etcd
func Init(endpoints []string) *etcdv3.Client {
	once.Do(func() {
		var err error
		if client, err = etcdv3.New(etcdv3.Config{Endpoints: endpoints, DialTimeout: 3 * time.Second}); err != nil {
			slog.Error("Init Etcd Failed", "error", err)
		}
	})
	return client
}
