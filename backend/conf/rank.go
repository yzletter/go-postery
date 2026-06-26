package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type RankServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
}

// LoadRankServiceConfig 加载 RankService 的非公共配置。
func LoadRankServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) RankServiceConfig {
	config := RankServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
