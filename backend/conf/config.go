package conf

import (
	"context"
	"log/slog"
	"strconv"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

var (
	watchKeys = make(map[string]struct{})
)

// LoadCommonMicroConf 加载微服务公共配置
func LoadCommonMicroConf(ctx context.Context, client *etcdv3.Client, prefix string) CommonMicroConf {
	conf := CommonMicroConf{
		// 数据库
		MySQL: LoadMySQLConfig(ctx, client, prefix),
		// 缓存
		Redis: LoadRedisConfig(ctx, client, prefix),
		// 消息队列
		Kafka:    LoadKafkaConfig(ctx, client, prefix),
		RabbitMQ: LoadRabbitMQConfig(ctx, client, prefix),
		RocketMQ: LoadRocketMQConfig(ctx, client, prefix),
		Milvus:   LoadMilvusConfig(ctx, client, prefix),
		// 链路追踪
		Jaeger: LoadJaegerConfig(ctx, client, prefix),
		// 服务发现
		ServiceHub: LoadServiceHubConfig(ctx, client, prefix),
	}

	go WatchConfig(ctx, client, prefix)

	return conf
}

// LoadOSSConfig 加载 OSS 配置
func LoadOSSConfig(ctx context.Context, client *etcdv3.Client, prefix string) OSSConfig {
	var config OSSConfig

	if resp, err := client.Get(ctx, prefix+"oss_access_key_id"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AccessKeyID = string(resp.Kvs[0].Value)
			watchKeys[prefix+"oss_access_key_id"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"oss_access_key_secret"); err == nil {
		if len(resp.Kvs) > 0 {
			config.AccessKeySecret = string(resp.Kvs[0].Value)
			watchKeys[prefix+"oss_access_key_secret"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"oss_arn"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Arn = string(resp.Kvs[0].Value)
			watchKeys[prefix+"oss_arn"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"oss_region"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Region = string(resp.Kvs[0].Value)
			watchKeys[prefix+"oss_region"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"oss_bucket"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Bucket = string(resp.Kvs[0].Value)
			watchKeys[prefix+"oss_bucket"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"oss_callback_url"); err == nil {
		if len(resp.Kvs) > 0 {
			config.CallbackURL = string(resp.Kvs[0].Value)
			watchKeys[prefix+"oss_callback_url"] = struct{}{}
		}
	}

	return config
}

// LoadAppConfig 加载应用配置
func LoadAppConfig(ctx context.Context, client *etcdv3.Client, prefix string) AppConfig {
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

// LoadWSGatewayHTTPConfig 加载 WebSocket 网关 HTTP 配置
func LoadWSGatewayHTTPConfig(ctx context.Context, client *etcdv3.Client, prefix string) WSGatewayHTTPConfig {
	var config WSGatewayHTTPConfig

	if resp, err := client.Get(ctx, prefix+"http_port"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Port = string(resp.Kvs[0].Value)
			watchKeys[prefix+"http_port"] = struct{}{}
		}
	}

	return config
}

// LoadArkConfig 加载 Ark 配置
func LoadArkConfig(ctx context.Context, client *etcdv3.Client, prefix string) ArkConfig {
	var config ArkConfig

	if resp, err := client.Get(ctx, prefix+"ark_embedder_model"); err == nil {
		if len(resp.Kvs) > 0 {
			config.EmbedderModel = string(resp.Kvs[0].Value)
			watchKeys[prefix+"ark_embedder_model"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"ark_llm_model"); err == nil {
		if len(resp.Kvs) > 0 {
			config.LLMModel = string(resp.Kvs[0].Value)
			watchKeys[prefix+"ark_llm_model"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"ark_api_key"); err == nil {
		if len(resp.Kvs) > 0 {
			config.APIKey = string(resp.Kvs[0].Value)
			watchKeys[prefix+"ark_api_key"] = struct{}{}
		}
	}

	return config
}

// LoadQwenConfig 加载 Qwen 配置
func LoadQwenConfig(ctx context.Context, client *etcdv3.Client, prefix string) QwenConfig {
	var config QwenConfig

	if resp, err := client.Get(ctx, prefix+"qwen_embedder_model"); err == nil {
		if len(resp.Kvs) > 0 {
			config.EmbedderModel = string(resp.Kvs[0].Value)
			watchKeys[prefix+"ark_embedder_model"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"qwen_base_url"); err == nil {
		if len(resp.Kvs) > 0 {
			config.BaseURL = string(resp.Kvs[0].Value)
			watchKeys[prefix+"qwen_base_url"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"qwen_llm_model"); err == nil {
		if len(resp.Kvs) > 0 {
			config.LLMModel = string(resp.Kvs[0].Value)
			watchKeys[prefix+"ark_llm_model"] = struct{}{}
		}
	}

	if resp, err := client.Get(ctx, prefix+"qwen_api_key"); err == nil {
		if len(resp.Kvs) > 0 {
			config.APIKey = string(resp.Kvs[0].Value)
			watchKeys[prefix+"qwen_api_key"] = struct{}{}
		}
	}

	return config
}

// WatchConfig 监听配置变化
func WatchConfig(ctx context.Context, client *etcdv3.Client, prefix string) {
	ch := client.Watch(ctx, prefix, etcdv3.WithPrefix())
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

// LoadRedisConfig 加载 Redis 配置
func LoadRedisConfig(ctx context.Context, client *etcdv3.Client, prefix string) RedisConfig {
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

// LoadGithubConfig 加载 Github 配置
func LoadGithubConfig(ctx context.Context, client *etcdv3.Client, prefix string) GithubConfig {
	var config GithubConfig

	// 获取地址
	if resp, err := client.Get(ctx, prefix+"github_token"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Token = string(resp.Kvs[0].Value)
			watchKeys[prefix+"github_token"] = struct{}{}
		}
	}

	return config
}

// LoadMilvusConfig 加载 Milvus 配置
func LoadMilvusConfig(ctx context.Context, client *etcdv3.Client, prefix string) MilvusConfig {
	var config MilvusConfig

	// 获取地址
	if resp, err := client.Get(ctx, prefix+"milvus_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"milvus_addr"] = struct{}{}
		}
	}

	return config
}

// LoadPrometheusConfig 加载 Prometheus 配置
func LoadPrometheusConfig(ctx context.Context, client *etcdv3.Client, prefix string) MetricConfig {
	var config MetricConfig

	// 获取端口
	if resp, err := client.Get(ctx, prefix+"prometheus_port"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Port = string(resp.Kvs[0].Value)
			watchKeys[prefix+"prometheus_port"] = struct{}{}
		}
	}

	return config
}

// LoadJaegerConfig 加载 Jaeger 配置
func LoadJaegerConfig(ctx context.Context, client *etcdv3.Client, prefix string) JaegerConfig {
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

// LoadKafkaConfig 加载 Kafka 配置
func LoadKafkaConfig(ctx context.Context, client *etcdv3.Client, prefix string) KafkaConfig {
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

// LoadRabbitMQConfig 加载 RabbitMQ 配置
func LoadRabbitMQConfig(ctx context.Context, client *etcdv3.Client, prefix string) RabbitMQConfig {
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

// LoadRocketMQConfig 加载 RocketMQ 配置
func LoadRocketMQConfig(ctx context.Context, client *etcdv3.Client, prefix string) RocketMQConfig {
	var config RocketMQConfig

	// 获取 RocketMQ 端口
	if resp, err := client.Get(ctx, prefix+"rocket_addr"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Addr = string(resp.Kvs[0].Value)
			watchKeys[prefix+"rocket_addr"] = struct{}{}
		}
	}

	return config
}

// LoadMySQLConfig 加载 MySQL 配置
func LoadMySQLConfig(ctx context.Context, client *etcdv3.Client, prefix string) MySQLConfig {
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

// LoadEmailConfig 加载邮件配置
func LoadEmailConfig(ctx context.Context, client *etcdv3.Client, prefix string) EmailConfig {
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

// LoadSMSConfig 加载短信配置
func LoadSMSConfig(ctx context.Context, client *etcdv3.Client, prefix string) SMSConfig {
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

// LoadLogConfig 加载日志配置
func LoadLogConfig(ctx context.Context, client *etcdv3.Client, prefix string) LogConfig {
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

// LoadGRPCConfig 加载 gRPC 配置
func LoadGRPCConfig(ctx context.Context, client *etcdv3.Client, prefix string) GrpcConfig {
	var config GrpcConfig

	// 获取 gRPC 端口
	if resp, err := client.Get(ctx, prefix+"grpc_port"); err == nil {
		if len(resp.Kvs) > 0 {
			config.Port = string(resp.Kvs[0].Value)
			watchKeys[prefix+"grpc_port"] = struct{}{}
		}
	}

	return config
}

// LoadServiceHubConfig 加载服务注册配置
func LoadServiceHubConfig(ctx context.Context, client *etcdv3.Client, prefix string) ServiceHubConfig {
	var config ServiceHubConfig

	// 获取心跳频率
	if resp, err := client.Get(ctx, prefix+"service_hub_heartbeat_frequency"); err == nil {
		if len(resp.Kvs) > 0 {
			config.HeartbeatFrequency, _ = strconv.Atoi(string(resp.Kvs[0].Value))
			watchKeys[prefix+"service_hub_heartbeat_frequency"] = struct{}{}
		}
	}

	// 获取服务注册前缀
	if resp, err := client.Get(ctx, prefix+"service_hub_register_prefix"); err == nil {
		if len(resp.Kvs) > 0 {
			config.ServiceRegisterPrefix = string(resp.Kvs[0].Value)
			watchKeys[prefix+"service_hub_register_prefix"] = struct{}{}
		}
	}

	return config
}
