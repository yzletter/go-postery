package config

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

var (
	watchKeys = make(map[string]struct{})
	conce     sync.Once
	config    Config
)

// LoadGlobalConfig 加载远程配置
//
// client 依赖注入 etcd Client
//
// prefix 业务配置 Key 的前缀
func LoadGlobalConfig(ctx context.Context, client *etcdv3.Client, prefix string) Config {
	conce.Do(func() {
		// 加载 Redis 配置
		config.Redis = loadRedisConfig(ctx, client, prefix)
		// 加载 Metric 配置
		config.Metric = loadPrometheusConfig(ctx, client, prefix)
		// 加载 Jaeger 配置
		config.Jaeger = loadJaegerConfig(ctx, client, prefix)
		// 加载 gRPC 配置
		config.GRPC = loadGRPCConfig(ctx, client, prefix)
		// 加载 Email 配置
		config.Email = loadEmailConfig(ctx, client, prefix)
		// 加载 SMS 配置
		config.SMS = loadSMSConfig(ctx, client, prefix)
		// 加载 Log 配置
		config.Log = loadLogConfig(ctx, client, prefix)

		fmt.Println(config)

		go watch(ctx, client, prefix, watchKeys)
	})

	return config
}

// 监听配置变化
func watch(ctx context.Context, client *etcdv3.Client, prefix string, keys map[string]struct{}) {
	ch := client.Watch(ctx, prefix, etcdv3.WithPrefix())
	for resp := range ch {
		for _, event := range resp.Events {
			// 只关心 PUT 即数据更新
			if event.Type == etcdv3.EventTypePut {
				key := string(event.Kv.Key)
				value := string(event.Kv.Value)
				if _, exists := keys[key]; exists {
					slog.Info("Config Has Changed", "name", key, "change", value)
				}
			}
		}
	}
}

func loadRedisConfig(ctx context.Context, client *etcdv3.Client, prefix string) RedisConfig {
	var config RedisConfig

	// 获取地址
	if resp, err := client.Get(ctx, prefix+"redis_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"redis_addr"] = struct{}{}
		}
	}

	// 获取数据库号
	if resp, err := client.Get(ctx, prefix+"redis_db"); err == nil {
		if len(resp.Kvs) > 0 {
			config.DB, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[prefix+"redis_db"] = struct{}{}
		}
	}

	return config
}

func loadPrometheusConfig(ctx context.Context, client *etcdv3.Client, prefix string) MetricConfig {
	var config MetricConfig

	// 获取地址
	if resp, err := client.Get(ctx, prefix+"prometheus_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"prometheus_addr"] = struct{}{}
		}
	}

	return config
}

func loadJaegerConfig(ctx context.Context, client *etcdv3.Client, prefix string) JaegerConfig {
	var config JaegerConfig

	// 获取地址
	if resp, err := client.Get(ctx, prefix+"jaeger_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"jaeger_addr"] = struct{}{}
		}
	}

	return config
}

func loadEmailConfig(ctx context.Context, client *etcdv3.Client, prefix string) EmailConfig {
	var config EmailConfig

	// 获取发信方
	if resp, err := client.Get(ctx, prefix+"email_from"); err == nil {
		if len(resp.Kvs) > 0 {
			config.From = string(resp.Kvs[0].Value)
			watchKeys[prefix+"email_from"] = struct{}{}
		}
	}

	// 获取授权码
	if resp, err := client.Get(ctx, prefix+"email_auth_code"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AuthCode = string(resp.Kvs[0].Value)
			watchKeys[prefix+"email_auth_code"] = struct{}{}
		}
	}

	// 获取主题
	if resp, err := client.Get(ctx, prefix+"email_subject"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Subject = string(resp.Kvs[0].Value)
			watchKeys[prefix+"email_subject"] = struct{}{}
		}
	}

	// 获取应用名称
	if resp, err := client.Get(ctx, prefix+"email_app_name"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AppName = string(resp.Kvs[0].Value)
			watchKeys[prefix+"email_app_name"] = struct{}{}
		}
	}

	// 获取有效时间
	if resp, err := client.Get(ctx, prefix+"email_expire_min"); err == nil {
		if len(resp.Kvs) > 0 {
			config.ExpireMin, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[prefix+"email_expire_min"] = struct{}{}
		}
	}

	// 获取年份
	if resp, err := client.Get(ctx, prefix+"email_year"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Year, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[prefix+"email_year"] = struct{}{}
		}
	}

	// 获取公司地址
	if resp, err := client.Get(ctx, prefix+"email_address"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Address = string(resp.Kvs[0].Value)
			watchKeys[prefix+"email_address"] = struct{}{}
		}
	}

	return config
}

func loadSMSConfig(ctx context.Context, client *etcdv3.Client, prefix string) SMSConfig {
	var config SMSConfig

	// 获取 AccessKeyID
	if resp, err := client.Get(ctx, prefix+"sms_access_key_id"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AccessKeyID = string(resp.Kvs[0].Value)
			watchKeys[prefix+"sms_access_key_id"] = struct{}{}
		}
	}

	// 获取 AccessKeySecret
	if resp, err := client.Get(ctx, prefix+"sms_access_key_secret"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AccessKeySecret = string(resp.Kvs[0].Value)
			watchKeys[prefix+"sms_access_key_secret"] = struct{}{}
		}
	}

	return config
}

func loadLogConfig(ctx context.Context, client *etcdv3.Client, prefix string) LogConfig {
	var config LogConfig

	// 获取日志文件路径
	if resp, err := client.Get(ctx, prefix+"log_filepath"); err == nil {
		if len(resp.Kvs) > 0 {
			config.FilePath = string(resp.Kvs[0].Value)
			watchKeys[prefix+"log_filepath"] = struct{}{}
		}
	}

	return config
}

func loadGRPCConfig(ctx context.Context, client *etcdv3.Client, prefix string) GRPCConfig {
	var config GRPCConfig

	// 获取 gRPC 端口
	if resp, err := client.Get(ctx, prefix+"grpc_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"grpc_addr"] = struct{}{}
		}
	}

	return config
}
