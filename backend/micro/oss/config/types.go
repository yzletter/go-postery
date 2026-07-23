package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// OSSServiceConfig OSS 微服务私有配置
type OSSServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	OSS    conf.OSSConfig
	GRPC   conf.GrpcConfig
}

// LoadOSSServiceConfig 加载 OSS 微服务私有配置
func LoadOSSServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) OSSServiceConfig {
	config := OSSServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		OSS:    conf.LoadOSSConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
