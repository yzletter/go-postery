package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	SearchKafkaTopic = "index_search"
	SearchKafkaGroup = "search_index"
)

type SearchServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
}

func LoadSearchServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) SearchServiceConfig {
	config := SearchServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
