package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	KeyUserScore   = "user:score"
	UserKafkaTopic = "follow"
	UserKafkaGroup = "follow"
)

type UserServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	OSS    OSSConfig
	GRPC   GRPCConfig
}

func LoadUserServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) UserServiceConfig {
	config := UserServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		OSS:    loadOSSConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
