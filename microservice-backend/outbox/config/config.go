package config

import (
	"context"
	"fmt"
	"log/slog"
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
		// 加载 MySQL 配置
		config.MySQL = loadMySQLConfig(ctx, client, prefix)
		// 加载 Kafka 配置
		config.Kafka = loadKafkaConfig(ctx, client, prefix)
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
