package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// BFFServiceConfig BFF 微服务私有配置
type BFFServiceConfig struct {
	App conf.AppConfig
	Log conf.LogConfig
}

// LoadBFFServiceConfig 加载 BFF 微服务私有配置
func LoadBFFServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) BFFServiceConfig {
	config := BFFServiceConfig{
		App: conf.LoadAppConfig(ctx, client, prefix),
		Log: conf.LoadLogConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
