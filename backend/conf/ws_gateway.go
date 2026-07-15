package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type WSGatewayServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
	HTTP   WSGatewayHTTPConfig
}

func LoadWSGatewayServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) WSGatewayServiceConfig {
	config := WSGatewayServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
		HTTP:   loadWSGatewayHTTPConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}

func loadWSGatewayHTTPConfig(ctx context.Context, client *etcdv3.Client, prefix string) WSGatewayHTTPConfig {
	var config WSGatewayHTTPConfig

	if resp, err := client.Get(ctx, prefix+"http_port"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Port = string(resp.Kvs[0].Value)
			watchKeys[prefix+"http_port"] = struct{}{}
		}
	}

	return config
}
