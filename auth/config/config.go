package config

import (
	"context"
	"strconv"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

const ConfigPrefix = "auth_service_"

var watchKeys = make(map[string]struct{})

func LoadGlobalConfig(ctx context.Context, client *etcdv3.Client) Config {
	var config Config
	// 加载 Redis 配置
	config.Redis = loadRedisConfig(ctx, client)
	// 加载 MySQL 配置
	config.MySQL = loadMySQLConfig(ctx, client)
	// 加载 Log 配置
	config.Log = loadLogConfig(ctx, client)
	// 加载 Metric 配置
	config.Metric = loadPrometheusConfig(ctx, client)
	// 加载 gRPC 配置
	config.GRPC = loadGRPCConfig(ctx, client)

	return config
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

func loadMySQLConfig(ctx context.Context, client *etcdv3.Client) MySQLConfig {
	var config MySQLConfig

	// 获取 Addr
	if resp, err := client.Get(ctx, ConfigPrefix+"MySQL_Addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"MySQL_Host"] = struct{}{}
		}
	}

	// 获取 User
	if resp, err := client.Get(ctx, ConfigPrefix+"MySQL_User"); err == nil {
		if len(resp.Kvs) > 0 {
			config.User = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"MySQL_User"] = struct{}{}
		}
	}

	// 获取 Password
	if resp, err := client.Get(ctx, ConfigPrefix+"MySQL_Password"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Password = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"MySQL_Password"] = struct{}{}
		}
	}

	// 获取 DBName
	if resp, err := client.Get(ctx, ConfigPrefix+"MySQL_DBName"); err == nil {
		if len(resp.Kvs) > 0 {
			config.DBName = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"MySQL_DBName"] = struct{}{}
		}
	}

	// 获取 LogFileName
	if resp, err := client.Get(ctx, ConfigPrefix+"MySQL_LogFilePath"); err == nil {
		if len(resp.Kvs) > 0 {
			config.LogFilePath = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"MySQL_LogFileName"] = struct{}{}
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

func loadGRPCConfig(ctx context.Context, client *etcdv3.Client) GRPCConfig {
	var config GRPCConfig

	// 获取 gRPC 端口
	if resp, err := client.Get(ctx, ConfigPrefix+"GRPC_Port"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Port = string(resp.Kvs[0].Value)
			watchKeys[ConfigPrefix+"GRPC_Port"] = struct{}{}
		}
	}

	return config
}
