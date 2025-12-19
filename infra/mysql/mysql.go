package infra

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/yzletter/go-postery/infra/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	globalDB *gorm.DB
)

type PoolConfig struct {
	MaxIdleConns    int           // 空闲连接数目上限
	MaxOpenConns    int           // 最多打开链接数目上限
	ConnMaxLifetime time.Duration // 单个连接的连接时间上限
}

// Init 初始化数据库
func Init(confDir, confFileName, confFileType, logDir string) *gorm.DB {
	// 读取 MySQL 相关配置
	vip := viper.InitViper(confDir, confFileName, confFileType) // 初始化一个 Viper 进行配置读取
	host := vip.GetString("mysql.host")
	port := vip.GetInt("mysql.port")
	user := vip.GetString("mysql.user")
	password := vip.GetString("mysql.password")
	dbName := vip.GetString("mysql.dbName")
	logFileName := vip.GetString("mysql.logFileName")

	// 拼接出 MySQL DataSourceName
	dataSourceName := getDataSourceName(user, password, host, port, dbName)

	// 设置 logger 相关配置
	loggerConfig := logger.Config{
		SlowThreshold:             100 * time.Millisecond, // 超过此阈值为慢查询
		Colorful:                  false,                  // 禁用颜色, 可提高性能
		IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 这种错误日志
		LogLevel:                  logger.Info,            // 日志最低阈值
	}
	// 初始化 MySQl Logger
	DBlogger := initDBLogger(logDir, logFileName, loggerConfig)

	// 设置 gorm 相关配置
	gormConfig := &gorm.Config{
		PrepareStmt:            true, // 执行任一 SQL 语句时, 都会创建 Prepare Statement 并缓存, 以提高后续执行效率
		SkipDefaultTransaction: true, // 禁止在事务中进行写入操作, 性能提升约 30%
		// 覆盖默认命名策略
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 表名映射不加复数, 仅仅是驼峰转为蛇形
		},
		Logger: DBlogger, // 日志控制
	}

	// 建立 MySQL 连接
	db, err := gorm.Open(mysql.Open(dataSourceName), gormConfig) // 生成一个 *gorm.DB
	if err != nil {
		slog.Error("初始化 MySQL 失败 ...", "error", err)
		panic(err)
	}
	slog.Info("初始化 MySQL 成功 ...")

	// 连接池配置
	poolConfig := PoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    10,
		ConnMaxLifetime: time.Hour,
	}
	configureDBPool(db, poolConfig)
	slog.Info("配置 MySQL 连接池成功 ...")

	// 赋给全局变量 globalDB
	globalDB = db

	return globalDB
}

// Ping 保持与 MySQL 的连接
func Ping() {
	if globalDB != nil {
		sqlDB, _ := globalDB.DB()
		err := sqlDB.Ping()
		if err != nil {
			slog.Info("Ping MySQL 失败 ...")
			return
		}
		slog.Info("Ping MySQL 成功 ...")
		return
	}
}

// Close 关闭 MySQL 连接
func Close() {
	if globalDB != nil {
		sqlDB, _ := globalDB.DB()
		err := sqlDB.Close()
		if err != nil {
			slog.Info("关闭 MySQL 失败 ...")
			return
		}
		slog.Info("关闭 MySQL 成功 ...")
		return
	}
}

/*
	InternalFunctions 内部函数
*/

// 拼接出 MySQL DataSourceName
func getDataSourceName(user string, password string, host string, port int, dbName string) string {
	// 拼接完整的请求路径 user:password@tcp(host:port)/dbName?charset=utf8mb4&parseTime=True&loc=Local
	// 使用 UTF-8mb4 编码, 解析时间为 Go 语言的时间类型, 按系统时区解析时间字段
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbName)
}

// 初始化 MySQL 日志
func initDBLogger(logDir string, logFileName string, loggerConfig logger.Config) logger.Interface {
	// 打开 logger 文件
	logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		slog.Info("打开 MySQL logger 文件失败 ...")
		panic(err)
	}

	// 返回 logger
	return logger.New(
		log.New(logFile, "\r\n", log.LstdFlags), // 每条 message 的前面都加上 \r\n, message 自动包含日期和时间
		loggerConfig,
	)
}

// 设置 MySQL 连接池参数
func configureDBPool(db *gorm.DB, poolConfig PoolConfig) {
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	// 超时会自动关闭, 因为数据库本身可能也对NoActive连接设置了超时时间, 我们的应对办法: 定期 ping, 或者 SetConnMaxLifetime
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
}
