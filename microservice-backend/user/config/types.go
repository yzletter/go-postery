package config

// Config 需要的所有配置
type Config struct {
	Redis  RedisConfig
	MySQL  MySQLConfig
	Kafka  KafkaConfig
	Jaeger JaegerConfig
	Log    LogConfig
	Metric MetricConfig
	OSS    OSSConfig
	GRPC   GRPCConfig
}

type OSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Arn             string
}

type RedisConfig struct {
	Addr string
	DB   int
}

type MySQLConfig struct {
	Addr        string
	User        string
	Password    string
	DBName      string
	LogFileDir  string
	LogFilename string
}

type MetricConfig struct {
	Addr string
}

type JaegerConfig struct {
	Addr string // Jaeger 地址
}

type KafkaConfig struct {
	Addr string
}

type LogConfig struct {
	FilePath string
}

type GRPCConfig struct {
	Addr string
}
