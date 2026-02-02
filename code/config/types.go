package config

type Config struct {
	Redis  RedisConfig
	Metric MetricConfig
	GRPC   GRPCConfig
	Email  EmailConfig
	SMS    SMSConfig
	Log    LogConfig
}

type RedisConfig struct {
	Addr string
	DB   int
}

type MetricConfig struct {
	Addr string
}

type EmailConfig struct {
	From      string // 发信方
	AuthCode  string // 授权码
	Subject   string // 主题
	AppName   string // 应用名称
	ExpireMin int    // 有效时间
	Year      int    // 年份
	Address   string // 公司地址
}

type SMSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
}

type LogConfig struct {
	FilePath string
}

type GRPCConfig struct {
	Addr string
}
