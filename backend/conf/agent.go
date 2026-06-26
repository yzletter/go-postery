package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	AgentQdrantKafkaTopic = "upsert_qdrant"
	AgentQdrantKafkaGroup = "agent_qdrant"
	AgentKafkaTopic       = "index_document"
	AgentKafkaGroup       = "agent_document"
)

type AgentServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
	Ark    ArkConfig
}

func LoadAgentServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) AgentServiceConfig {
	config := AgentServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
		Ark:    loadArkConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
