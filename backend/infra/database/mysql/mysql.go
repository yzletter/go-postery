package infra

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"sync"
	"time"

	"github.com/yzletter/go-postery/backend/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	globalDB *gorm.DB
	dbOnce   sync.Once
)

type PoolConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

func Init(config conf.MySQLConfig) *gorm.DB {
	dbOnce.Do(func() {
		dataSourceName := getDataSourceName(config.User, config.Password, config.Addr, config.DBName)

		// MySQL 日志配置
		loggerConfig := logger.Config{
			ParameterizedQueries:      false,                  // true 代表 SQL 日志里不包含参数
			SlowThreshold:             100 * time.Millisecond, // 耗时超过此值认定为慢查询
			Colorful:                  false,                  // 禁用颜色
			IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 这种错误日志
			LogLevel:                  logger.Info,            // LogLevel的最低阈值，Silent为不输出日志
		}

		// 初始化 MySQL 日志
		dbLogger := initDBLogger(config.LogFileDir, config.LogFilename, loggerConfig)

		// Gorm 配置
		gormConfig := &gorm.Config{
			PrepareStmt:              true,     // 执行任何 SQL 时都会创建一个 prepared statement 并将其缓存，以提高后续的效率
			SkipDefaultTransaction:   true,     // 为了确保数据一致性，GORM 会在事务里执行写入操作（创建、更新、删除）如果没有这方面的要求，可以在初始化时禁用它，这将获得大约 30%+ 性能提升。
			Logger:                   dbLogger, // 日志
			DryRun:                   false,    // true 代表生成 SQL 但不执行，可以用于准备或测试生成的 SQL
			DisableNestedTransaction: true,     // 在一个事务中使用 Transaction 方法，GORM会使用 SavePoint(savedPointName)，RollbackTo(savedPointName) 为你提供嵌套事务支持。如果不需要嵌套事务，可以将其禁用
			DisableAutomaticPing:     true,     // 在完成初始化后，GORM 会自动 ping 数据库以检查数据库的可用性
			NamingStrategy: schema.NamingStrategy{ // 覆盖默认的 NamingStrategy 来更改命名约定
				SingularTable: false, // true 表示表名映射时不加复数，仅是驼峰 --> 蛇形
			},
		}

		// 连接数据库
		db, err := gorm.Open(mysql.Open(dataSourceName), gormConfig)
		if err != nil {
			slog.Error("init MySQL failed ...", "error", err)
			panic(err)
		}
		slog.Info("init MySQL success ...")

		// 配置连接池
		configureDBPool(db, PoolConfig{
			MaxIdleConns:    10,
			MaxOpenConns:    10,
			ConnMaxLifetime: time.Hour,
		})

		globalDB = db
	})
	return globalDB
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
	// 连接池控制参数
	sqlDB, _ := db.DB()
	// 池子里空闲连接的数量上限（超出此上限就把相应的连接关闭掉）
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	// 最多开这么多连接
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	// 一个连接最多可使用这么长时间，超时后连接会自动关闭（因为数据库本身可能也对NoActive连接设置了超时时间，我们的应对办法：定期ping，或者SetConnMaxLifetime）
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
}
