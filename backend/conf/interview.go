package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

// InterviewServiceConfig 面试 Agent 私有配置
type InterviewServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
	Ark    ArkConfig
	Qwen   QwenConfig
	Github GithubConfig
}

func LoadInterviewServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) InterviewServiceConfig {
	config := InterviewServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
		Ark:    loadArkConfig(ctx, client, prefix),
		Qwen:   loadQwenConfig(ctx, client, prefix),
		Github: loadGithubConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
