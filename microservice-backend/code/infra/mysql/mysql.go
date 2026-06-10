package infra

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/yzletter/go-postery/microservice-backend/code/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	globalDB *gorm.DB
)

type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func Init(config conf.MySQLConfig) *gorm.DB {
	dataSourceName := getDataSourceName(config.User, config.Password, config.Addr, config.DBName)

	loggerConfig := logger.Config{
		SlowThreshold:             100 * time.Millisecond,
		Colorful:                  false,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  logger.Info,
	}
	dbLogger := initDBLogger(config.LogFileDir, config.LogFilename, loggerConfig)

	gormConfig := &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
		Logger: dbLogger,
	}

	db, err := gorm.Open(mysql.Open(dataSourceName), gormConfig)
	if err != nil {
		slog.Error("初始化 MySQL 失败 ...", "error", err)
		panic(err)
	}
	slog.Info("初始化 MySQL 成功 ...")

	configureDBPool(db, PoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Hour,
	})

	globalDB = db
	return globalDB
}

func Ping() {
	if globalDB == nil {
		return
	}
	sqlDB, _ := globalDB.DB()
	if err := sqlDB.Ping(); err != nil {
		slog.Info("Ping MySQL 失败 ...")
		return
	}
	slog.Info("Ping MySQL 成功 ...")
}

func Close() {
	if globalDB == nil {
		return
	}
	sqlDB, _ := globalDB.DB()
	if err := sqlDB.Close(); err != nil {
		slog.Info("关闭 MySQL 失败 ...")
		return
	}
	slog.Info("关闭 MySQL 成功 ...")
}

func getDataSourceName(user string, password string, addr string, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, addr, dbName)
}

func initDBLogger(logDir string, logFileName string, loggerConfig logger.Config) logger.Interface {
	_ = os.MkdirAll(logDir, os.ModePerm)
	logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return logger.New(
		log.New(logFile, "\r\n", log.LstdFlags),
		loggerConfig,
	)
}

func configureDBPool(db *gorm.DB, poolConfig PoolConfig) {
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
}
