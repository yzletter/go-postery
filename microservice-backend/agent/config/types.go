package config

// Config 需要的所有配置
type Config struct {
	CommonMicroServiceConfig
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
	Ark    ArkConfig
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

type ArkConfig struct {
	EmbedderModel string
	LLMModel      string
	APIKey        string
}

type QdrantConfig struct {
	Host string
	Port int
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
