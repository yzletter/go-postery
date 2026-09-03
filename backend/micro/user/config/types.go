package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// UserServiceConfig User 微服务私有配置
type UserServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
}

// LoadUserServiceConfig 加载 User 微服务私有配置
func LoadUserServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) UserServiceConfig {
	config := UserServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
