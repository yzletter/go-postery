package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// CodeServiceConfig Code 微服务私有配置
type CodeServiceConfig struct {
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
	Email  conf.EmailConfig
	SMS    conf.SMSConfig
	Log    conf.LogConfig
}

// LoadCodeServiceConfig 加载 Code 微服务私有配置
func LoadCodeServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) CodeServiceConfig {
	config := CodeServiceConfig{
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
		Email:  conf.LoadEmailConfig(ctx, client, prefix),
		SMS:    conf.LoadSMSConfig(ctx, client, prefix),
		Log:    conf.LoadLogConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
