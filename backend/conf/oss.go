package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

// OSSServiceConfig OSS 公共微服务私有配置
type OSSServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	OSS    OSSConfig
	GRPC   GrpcConfig
}

// LoadOSSServiceConfig 加载 OSS 公共微服务私有配置
func LoadOSSServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) OSSServiceConfig {
	config := OSSServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		OSS:    loadOSSConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
