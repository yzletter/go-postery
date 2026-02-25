package config

type Config struct {
	Redis  RedisConfig
	MySQL  MySQLConfig
	Log    LogConfig
	Metric MetricConfig
	GRPC   GRPCConfig
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
	LogFilePath string
}

type LogConfig struct {
	FilePath string
}

type MetricConfig struct {
	Addr string
}

type GRPCConfig struct {
	Port string
}
