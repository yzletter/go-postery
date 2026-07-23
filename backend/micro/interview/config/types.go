package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// InterviewServiceConfig 面试 Agent 私有配置
type InterviewServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
	Ark    conf.ArkConfig
	Qwen   conf.QwenConfig
	Github conf.GithubConfig
}

// LoadInterviewServiceConfig 加载面试 Agent 微服务私有配置
func LoadInterviewServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) InterviewServiceConfig {
	config := InterviewServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
		Ark:    conf.LoadArkConfig(ctx, client, prefix),
		Qwen:   conf.LoadQwenConfig(ctx, client, prefix),
		Github: conf.LoadGithubConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
