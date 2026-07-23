package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// SearchServiceConfig Search 微服务私有配置
type SearchServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
}

// LoadSearchServiceConfig 加载 Search 微服务私有配置
func LoadSearchServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) SearchServiceConfig {
	config := SearchServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
