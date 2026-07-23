package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// OutboxServiceConfig Outbox 微服务私有配置
type OutboxServiceConfig struct {
	Log conf.LogConfig
}

// LoadOutboxServiceConfig 加载 Outbox 微服务私有配置
func LoadOutboxServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) OutboxServiceConfig {
	config := OutboxServiceConfig{
		Log: conf.LoadLogConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
