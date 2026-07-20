package conf

type CommonMicroConf struct {
	MySQL      MySQLConfig
	Redis      RedisConfig
	Kafka      KafkaConfig
	RabbitMQ   RabbitMQConfig
	RocketMQ   RocketMQConfig
	Milvus     MilvusConfig
	Jaeger     JaegerConfig
	ServiceHub ServiceHubConfig
}

type AppConfig struct {
	FrontendAddr string
	BackendAddr  string
}

type RedisConfig struct {
	Addr string
	DB   int
}

type MilvusConfig struct {
	Addr string
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

type OSSConfig struct {
	AccessKeyID     string
	AccessKeySecret string
	Arn             string
	Region          string
	Bucket          string
	CallbackURL     string
}

type ArkConfig struct {
	EmbedderModel string
	LLMModel      string
	APIKey        string
}

type GithubConfig struct {
	Token string
}

type QwenConfig struct {
	BaseURL       string
	EmbedderModel string
	LLMModel      string
	APIKey        string
}

type MetricConfig struct {
	Port string
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

type WSGatewayHTTPConfig struct {
	Port string
}

type LogConfig struct {
	FilePath string
}

type GrpcConfig struct {
	Port string
}

type ServiceHubConfig struct {
	HeartbeatFrequency    int
	ServiceRegisterPrefix string
}
