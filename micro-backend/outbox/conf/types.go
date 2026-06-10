package conf

// Config 需要的所有配置
type Config struct {
	CommonMicroServiceConfig
	Log LogConfig
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

type MySQLConfig struct {
	Addr        string
	User        string
	Password    string
	DBName      string
	LogFileDir  string
	LogFilename string
}

type KafkaConfig struct {
	Addr string
}

type RedisConfig struct {
	Addr string
	DB   int
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

type JaegerConfig struct {
	Addr string // Jaeger 地址
}

type LogConfig struct {
	FilePath string
}
