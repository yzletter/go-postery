package config

// Config 需要的所有配置
type Config struct {
	MySQL MySQLConfig
	Kafka KafkaConfig
	Log   LogConfig
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

type LogConfig struct {
	FilePath string
}
