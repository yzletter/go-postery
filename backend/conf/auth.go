package conf

import (
	"context"

	"github.com/yzletter/go-postery/backend/micro/code/domain"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

const (
	UserIDInContext = "user_id" // uid 在上下文中的 Name
)

type CodeBiz = domain.BizType

const (
	CodeBizSMS   CodeBiz = domain.BizSMS
	CodeBizEmail CodeBiz = domain.BizEmail
)

type AuthServiceConfig struct {
	Log    LogConfig
	Metric MetricConfig
	GRPC   GrpcConfig
}

// LoadAuthServiceConfig 加载 AuthService 的非公共配置。
func LoadAuthServiceConfig(ctx context.Context, client *etcdv3.Client, prefix string) AuthServiceConfig {
	config := AuthServiceConfig{
		Log:    loadLogConfig(ctx, client, prefix),
		Metric: loadPrometheusConfig(ctx, client, prefix),
		GRPC:   loadGRPCConfig(ctx, client, prefix),
	}

	go watch(ctx, client, prefix, watchKeys)

	return config
}
