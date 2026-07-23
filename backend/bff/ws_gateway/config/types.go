package config

import (
	"context"

	"github.com/yzletter/go-postery/backend/conf"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// WSGatewayServiceConfig WebSocket 网关微服务私有配置
type WSGatewayServiceConfig struct {
	Log    conf.LogConfig
	Metric conf.MetricConfig
	GRPC   conf.GrpcConfig
	HTTP   conf.WSGatewayHTTPConfig
}

// LoadWSGatewayServiceConfig 加载 WebSocket 网关微服务私有配置
func LoadWSGatewayServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) WSGatewayServiceConfig {
	config := WSGatewayServiceConfig{
		Log:    conf.LoadLogConfig(ctx, client, prefix),
		Metric: conf.LoadPrometheusConfig(ctx, client, prefix),
		GRPC:   conf.LoadGRPCConfig(ctx, client, prefix),
		HTTP:   conf.LoadWSGatewayHTTPConfig(ctx, client, prefix),
	}

	go conf.WatchConfig(ctx, client, prefix)

	return config
}
