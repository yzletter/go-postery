package config

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const ConfigPrefix = "code_service_"

var watchKeys = make(map[string]struct{})
var conce sync.Once
var config Config

func LoadGlobalConfig(ctx context.Context, client *etcdv3.Client) Config {
	conce.Do(func() {
		// 加载 Redis 配置
		config.Redis = loadRedisConfig(ctx, client)
		// 加载 Metric 配置
		config.Metric = loadPrometheusConfig(ctx, client)
		// 加载 gRPC 配置
		config.GRPC = loadGRPCConfig(ctx, client)
		// 加载 Email 配置
		config.Email = loadEmailConfig(ctx, client)
		// 加载 SMS 配置
		config.SMS = loadSMSConfig(ctx, client)
		// 加载 Log 配置
		config.Log = loadLogConfig(ctx, client)

		fmt.Println(config)

		go watch(ctx, client, watchKeys)
	})

	return config
}

func watch(ctx context.Context, client *etcdv3.Client, watchKeys map[string]struct{}) {
	ch := client.Watch(ctx, ConfigPrefix, etcdv3.WithPrefix())
	for resp := range ch {
		for _, event := range resp.Events {
			// 只关心 PUT 即数据更新
			if event.Type == etcdv3.EventTypePut {
				key := string(event.Kv.Key)
				value := string(event.Kv.Value)
				if _, exists := watchKeys[key]; exists {
					slog.Info("Config Has Changed", "name", key, "change", value)
				}
			}
		}
	}
}

func loadRedisConfig(ctx context.Context, client *etcdv3.Client) RedisConfig {
	var config RedisConfig

	// 获取地址
	if resp, err := client.Get(ctx, ConfigPrefix+"Redis_Addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Redis_Addr"] = struct{}{}
		}
	}

	// 获取数据库号
	if resp, err := client.Get(ctx, ConfigPrefix+"Redis_DB"); err == nil {
		if len(resp.Kvs) > 0 {
			config.DB, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[ConfigPrefix+"Redis_DB"] = struct{}{}
		}
	}

	return config
}

func loadPrometheusConfig(ctx context.Context, client *etcdv3.Client) MetricConfig {
	var config MetricConfig

	// 获取地址
	if resp, err := client.Get(ctx, ConfigPrefix+"Prometheus_Addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Prometheus_Addr"] = struct{}{}
		}
	}

	return config
}

func loadEmailConfig(ctx context.Context, client *etcdv3.Client) EmailConfig {
	var config EmailConfig

	// 获取发信方
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_From"); err == nil {
		if len(resp.Kvs) > 0 {
			config.From = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Email_From"] = struct{}{}
		}
	}

	// 获取授权码
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_AuthCode"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AuthCode = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Email_AuthCode"] = struct{}{}
		}
	}

	// 获取主题
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_Subject"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Subject = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Email_Subject"] = struct{}{}
		}
	}

	// 获取应用名称
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_AppName"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AppName = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Email_AppName"] = struct{}{}
		}
	}

	// 获取有效时间
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_ExpireMin"); err == nil {
		if len(resp.Kvs) > 0 {
			config.ExpireMin, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[ConfigPrefix+"Email_ExpireMin"] = struct{}{}
		}
	}

	// 获取年份
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_Year"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Year, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[ConfigPrefix+"Email_Year"] = struct{}{}
		}
	}

	// 获取公司地址
	if resp, err := client.Get(ctx, ConfigPrefix+"Email_Address"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Address = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Email_Address"] = struct{}{}
		}
	}

	return config
}

func loadSMSConfig(ctx context.Context, client *etcdv3.Client) SMSConfig {
	var config SMSConfig

	// 获取 AccessKeyID
	if resp, err := client.Get(ctx, ConfigPrefix+"SMS_AccessKeyID"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AccessKeyID = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"SMS_AccessKeyID"] = struct{}{}
		}
	}

	// 获取 AccessKeySecret
	if resp, err := client.Get(ctx, ConfigPrefix+"SMS_AccessKeySecret"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AccessKeySecret = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"SMS_AccessKeySecret"] = struct{}{}
		}
	}

	return config
}

func loadLogConfig(ctx context.Context, client *etcdv3.Client) LogConfig {
	var config LogConfig

	// 获取日志文件路径
	if resp, err := client.Get(ctx, ConfigPrefix+"Log_FilePath"); err == nil {
		if len(resp.Kvs) > 0 {
			config.FilePath = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"Log_FilePath"] = struct{}{}
		}
	}

	return config
}

func loadGRPCConfig(ctx context.Context, client *etcdv3.Client) GRPCConfig {
	var config GRPCConfig

	// 获取 gRPC 端口
	if resp, err := client.Get(ctx, ConfigPrefix+"GRPC_Addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"GRPC_Addr"] = struct{}{}
		}
	}

	return config
}
