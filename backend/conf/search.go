package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	// SearchKafkaTopic 搜索索引消息 Topic
	SearchKafkaTopic = "index_search"
	// SearchKafkaGroup 搜索索引消费者组
	SearchKafkaGroup = "search_index"
)

type SearchServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
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
