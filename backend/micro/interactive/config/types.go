package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// InteractiveServiceConfig Interactive 微服务私有配置
type InteractiveServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
}

// LoadInteractiveServiceConfig 加载 Interactive 微服务私有配置
func LoadInteractiveServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) InteractiveServiceConfig {
	config := InteractiveServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
