package conf

import (
	"context"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	OutboxLockTime = 60
	OutboxInterval = 5 * time.Second
)

type OutboxServiceConfig struct {
	Log LogConfig
}

// LoadOutboxServiceConfig 加载 OutboxService 的非公共配置。
func LoadOutboxServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) OutboxServiceConfig {
	config := OutboxServiceConfig{
		Log: loadLogConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
