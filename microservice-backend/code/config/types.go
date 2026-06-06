package config

// Config 需要的所有配置
type Config struct {
	CommonMicroServiceConfig
	Metric MetricConfig
	GRPC   GRPCConfig
	Email  EmailConfig
	SMS    SMSConfig
	Log    LogConfig
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

type MetricConfig struct {
	Addr string
}

type JaegerConfig struct {
	Addr string // Jaeger 地址
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

type ServiceHubConfig struct {
	HeartbeatFrequency    int
	ServiceRegisterPrefix string
}
