package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// PostServiceConfig Post 微服务私有配置
type PostServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
}

// LoadPostServiceConfig 加载 Post 微服务私有配置
func LoadPostServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) PostServiceConfig {
	config := PostServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
