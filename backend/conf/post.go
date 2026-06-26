package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type PostServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
}

// LoadPostServiceConfig 加载 PostService 的非公共配置。
func LoadPostServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) PostServiceConfig {
	config := PostServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
