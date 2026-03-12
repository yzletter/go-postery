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
		// 加载 Jaeger 配置
		config.Jaeger = loadJaegerConfig(ctx, client, prefix)
		// 加载 App 配置
		config.App = loadAppConfig(ctx, client, prefix)
		// 加载 Log 配置
		config.Log = loadLogConfig(ctx, client, prefix)
		// 加载 RabbitMQ 配置
		config.RabbitMQ = loadRabbitMQConfig(ctx, client, prefix)
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

func loadAppConfig(ctx context.Context, client *etcdv3.Client, prefix string) AppConfig {
	var config AppConfig

	// 获取前端地址
	if resp, err := client.Get(ctx, prefix+"frontend_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.FrontendAddr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"frontend_addr"] = struct{}{}
		}
	}

	// 获取后端地址
	if resp, err := client.Get(ctx, prefix+"backend_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.BackendAddr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"backend_addr"] = struct{}{}
		}
	}

	return config
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

func loadKafkaConfig(ctx context.Context, client *etcdv3.Client, prefix string) KafkaConfig {
	var config KafkaConfig

	// 获取地址
	if resp, err := client.Get(ctx, prefix+"kafka_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"kafka_addr"] = struct{}{}
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

func loadMySQLConfig(ctx context.Context, client *etcdv3.Client, prefix string) MySQLConfig {
	var config MySQLConfig

	// 获取 Addr
	if resp, err := client.Get(ctx, prefix+"mysql_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"mysql_addr"] = struct{}{}
		}
	}

	// 获取 User
	if resp, err := client.Get(ctx, prefix+"mysql_user"); err == nil {
		if len(resp.Kvs) > 0 {
			config.User = string(resp.Kvs[0].Value)
			watchKeys[prefix+"mysql_user"] = struct{}{}
		}
	}

	// 获取 Password
	if resp, err := client.Get(ctx, prefix+"mysql_password"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Password = string(resp.Kvs[0].Value)
			watchKeys[prefix+"mysql_password"] = struct{}{}
		}
	}

	// 获取 DBName
	if resp, err := client.Get(ctx, prefix+"mysql_db_name"); err == nil {
		if len(resp.Kvs) > 0 {
			config.DBName = string(resp.Kvs[0].Value)
			watchKeys[prefix+"mysql_db_name"] = struct{}{}
		}
	}

	// 获取 LogFileDir
	if resp, err := client.Get(ctx, prefix+"mysql_log_file_dir"); err == nil {
		if len(resp.Kvs) > 0 {
			config.LogFileDir = string(resp.Kvs[0].Value)
			watchKeys[prefix+"mysql_log_file_dir"] = struct{}{}
		}
	}

	// 获取 LogFileName
	if resp, err := client.Get(ctx, prefix+"mysql_log_filename"); err == nil {
		if len(resp.Kvs) > 0 {
			config.LogFilename = string(resp.Kvs[0].Value)
			watchKeys[prefix+"mysql_log_filename"] = struct{}{}
		}
	}

	return config
}

func loadRabbitMQConfig(ctx context.Context, client *etcdv3.Client, prefix string) RabbitMQConfig {
	var config RabbitMQConfig

	// 获取 Addr
	if resp, err := client.Get(ctx, prefix+"rabbitmq_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"rabbitmq_addr"] = struct{}{}
		}
	}

	// 获取 User
	if resp, err := client.Get(ctx, prefix+"rabbitmq_user"); err == nil {
		if len(resp.Kvs) > 0 {
			config.User = string(resp.Kvs[0].Value)
			watchKeys[prefix+"rabbitmq_user"] = struct{}{}
		}
	}

	// 获取 Password
	if resp, err := client.Get(ctx, prefix+"rabbitmq_password"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Password = string(resp.Kvs[0].Value)
			watchKeys[prefix+"rabbitmq_password"] = struct{}{}
		}
	}

	return config
}
