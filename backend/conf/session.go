package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	SessionKafkaTopic = "session"
	SessionKafkaGroup = "session"
)

type SessionServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
}

func LoadSessionServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) SessionServiceConfig {
	config := SessionServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
