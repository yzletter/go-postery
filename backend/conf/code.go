package conf

import (
	"context"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

// 短信验证码
const (
	SendSMSInterval = 60            // 发送间隔
	SMSValidTime    = 300           // 有效时间
	PhoneCodePrefix = "phone:code:" // 前缀
)

// 邮箱验证码
const (
	SendEmailInterval = 60
	EmailValidTime    = 600
	EmailCodePrefix   = "email:code:"
)

type CodeServiceConfig struct {
	Metric MetricConfig
	GRPC   GRPCConfig
	Email  EmailConfig
	SMS    SMSConfig
	Log    LogConfig
}

// LoadCodeServiceConfig 加载 CodeService 的非公共配置。
func LoadCodeServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) CodeServiceConfig {
	config := CodeServiceConfig{
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
		Email:  loadEmailConfig(ctx, client, prefix),
		SMS:    loadSMSConfig(ctx, client, prefix),
		Log:    loadLogConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
