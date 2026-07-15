package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type AuthServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
}

// LoadAuthServiceConfig 加载 AuthService 的非公共配置。
func LoadAuthServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) AuthServiceConfig {
	config := AuthServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
