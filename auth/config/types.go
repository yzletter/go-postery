package config

type Config struct {
	Redis  RedisConfig
	MySQL  MySQLConfig
	Jaeger JaegerConfig
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
}

type JaegerConfig struct {
	Addr string // Jaeger 地址
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
