package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// LotteryServiceConfig Lottery 微服务私有配置
type LotteryServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
}

// LoadLotteryServiceConfig 加载 Lottery 微服务私有配置
func LoadLotteryServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) LotteryServiceConfig {
	config := LotteryServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
