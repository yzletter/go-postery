package conf

import "time"

const (
	RocketLotteryTopic             = "GO_POSTERY_CANCEL_ORDER"
	RocketLotteryConsumerGroup     = "go_postery"
	RocketAwaitDuration            = 5 * time.Second
	RocketLotteryPayDelay          = 600
	RocketLotteryInvisibleDuration = 10 * time.Second
)

// sh mqadmin updateTopic -n localhost:9876 -c DefaultCluster -t GO_POSTERY_CANCEL_ORDER -a +message.type=DELAY
// sh mqadmin deleteTopic -n localhost:9876 -c DefaultCluster -t GO_POSTERY_CANCEL_ORDER
// sh mqadmin updateSubGroup -n localhost:9876 -c DefaultCluster -g go_postery

type Config struct {
	Redis    RedisConfig
	MySQL    MySQLConfig
	Jaeger   JaegerConfig
	Log      LogConfig
	Metric   MetricConfig
	GRPC     GRPCConfig
	RocketMQ RocketMQConfig
}

type JaegerConfig struct {
	Addr string // Jaeger 地址
}

type RocketMQConfig struct {
	Addr string // RocketMQ 地址
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

type LogConfig struct {
	FilePath string
}

type MetricConfig struct {
	Addr string
}

type GRPCConfig struct {
	Addr string
}
