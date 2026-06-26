package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type InteractiveServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
}

// LoadInteractiveServiceConfig 加载 InteractiveService 的非公共配置
func LoadInteractiveServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) InteractiveServiceConfig {
	config := InteractiveServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
