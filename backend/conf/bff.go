package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

type AppConfig struct {
	FrontendAddr string
	BackendAddr  string
}

type BFFServiceConfig struct {
	App AppConfig
	Log LogConfig
}

func LoadBFFServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) BFFServiceConfig {
	config := BFFServiceConfig{
		App: loadAppConfig(ctx, client, prefix),
		Log: loadLogConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}

func loadAppConfig(ctx context.Context, client *etcdv3.Client, prefix string) AppConfig {
	var config AppConfig

	if resp, err := client.Get(ctx, prefix+"frontend_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.FrontendAddr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"frontend_addr"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"backend_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.BackendAddr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"backend_addr"] = struct{}{}
		}
	}

	return config
}
