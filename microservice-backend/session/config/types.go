package config

// Config 需要的所有配置
type Config struct {
	CommonMicroServiceConfig
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
}

type CommonMicroServiceConfig struct {
	MySQL    MySQLConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	RabbitMQ RabbitMQConfig
	RocketMQ RocketMQConfig
	Qdrant   QdrantConfig
	Jaeger   JaegerConfig
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
