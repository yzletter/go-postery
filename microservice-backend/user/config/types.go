package config

// Config 需要的所有配置
type Config struct {
	CommonMicroServiceConfig
	Log    LogConfig
	Metric MetricConfig
	OSS    OSSConfig
	GRPC   GRPCConfig
}

type CommonMicroServiceConfig struct {
	MySQL      MySQLConfig
	Redis      RedisConfig
	Kafka      KafkaConfig
	RabbitMQ   RabbitMQConfig
	RocketMQ   RocketMQConfig
	Qdrant     QdrantConfig
	Jaeger     JaegerConfig
	ServiceHub ServiceHubConfig
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
	Port string
}

type JaegerConfig struct {
	Addr string // Jaeger 地址
}

type KafkaConfig struct {
	Addr string
}

type RabbitMQConfig struct {
	User     string
	Password string
	Addr     string
}

type RocketMQConfig struct {
	Addr string // RocketMQ 地址
}

type QdrantConfig struct {
	Host string
	Port int
}

type LogConfig struct {
	FilePath string
}

type GRPCConfig struct {
	Port string
}

type ServiceHubConfig struct {
	HeartbeatFrequency    int
	ServiceRegisterPrefix string
}
