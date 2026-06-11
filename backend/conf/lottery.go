package conf

import (
	"context"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	RocketLotteryTopic             = "GO_POSTERY_CANCEL_ORDER"
	RocketLotteryConsumerGroup     = "go_postery"
	RocketAwaitDuration            = 5 * time.Second
	RocketLotteryPayDelay          = 600
	RocketLotteryInvisibleDuration = 10 * time.Second
)

type LotteryServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
}

// LoadLotteryServiceConfig 加载 LotteryService 的非公共配置。
func LoadLotteryServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) LotteryServiceConfig {
	config := LotteryServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
