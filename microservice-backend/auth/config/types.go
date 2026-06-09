package config

type Config struct {
	CommonMicroServiceConfig
	Log    LogConfig
	Metric MetricConfig
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

type JaegerConfig struct {
	Addr string // Jaeger 地址
}

type RedisConfig struct {
	Addr string
	DB   int
}

type KafkaConfig struct {
	Addr string // Kafka 地址
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

type MySQLConfig struct {
	Addr        string
	User        string
	Password    string
	DBName      string
	LogFileDir  string
	LogFilename string
}

type LogConfig struct {
	FilePath string
}

type MetricConfig struct {
	Port string
}

type GRPCConfig struct {
	Port string
}

type ServiceHubConfig struct {
	HeartbeatFrequency    int
	ServiceRegisterPrefix string
}
