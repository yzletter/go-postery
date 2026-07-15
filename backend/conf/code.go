package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type CodeServiceConfig struct {
	Metric MetricConfig
	GRPC   GrpcConfig
	Email  EmailConfig
	SMS    SMSConfig
	Log    LogConfig
}

// LoadCodeServiceConfig 加载 CodeService 的非公共配置。
func LoadCodeServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) CodeServiceConfig {
	config := CodeServiceConfig{
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
		Email:  loadEmailConfig(ctx, client, prefix),
		SMS:    loadSMSConfig(ctx, client, prefix),
		Log:    loadLogConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
