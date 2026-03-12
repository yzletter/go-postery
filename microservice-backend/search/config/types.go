package config

// Config 需要的所有配置
type Config struct {
	Metric MetricConfig
	Jaeger JaegerConfig
	Kafka  KafkaConfig
	GRPC   GRPCConfig
	Log    LogConfig
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
