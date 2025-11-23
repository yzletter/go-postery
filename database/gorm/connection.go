package database

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/yzletter/go-postery/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	GoPosteryDB *gorm.DB // 定义全局数据库变量 GoPosteryDB
)

// ConnectToDB 连接到 MySQL 数据库, 生成一个 *gorm.DB 赋给全局数据库变量 GoPosteryDB
func ConnectToDB(confDir, confFileName, confFileType, logDir string) {
	// 读取 MySQL 相关配置
	viper := utils.InitViper(confDir, confFileName, confFileType) // 初始化一个 Viper 进行配置读取
	host := viper.GetString("mysql.host")
	port := viper.GetInt("mysql.port")
	user := viper.GetString("mysql.user")
	password := viper.GetString("mysql.password")
	dbName := viper.GetString("mysql.dbName")
	logFileName := viper.GetString("mysql.logFileName")

	// 拼接完整的请求路径 user:password@tcp(host:port)/dbName?charset=utf8mb4&parseTime=True&loc=Local
	// 使用 UTF-8mb4 编码, 解析时间为 Go 语言的时间类型, 按系统时区解析时间字段
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbName)

	// 设置 logger 相关配置
	logFile, err := os.OpenFile(path.Join(logDir, logFileName), os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm) // 打开日志文件
	if err != nil {
		panic(fmt.Errorf("go-postery ConnectToDB : 打开目标 logger 文件失败 %s", err))
	}
	loggerConfig := logger.Config{ // logger 相关配置
		SlowThreshold:             100 * time.Millisecond, // 超过此阈值为慢查询
		Colorful:                  false,                  // 禁用颜色, 可提高性能
		IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 这种错误日志
		LogLevel:                  logger.Info,            // 日志最低阈值
	}
	myDBLogger := logger.New(
		log.New(logFile, "\r\n", log.LstdFlags), // 每条 message 的前面都加上 \r\n, message 自动包含日期和时间
		loggerConfig,
	)

	// 设置 gorm 相关配置
	gormConfig := &gorm.Config{
		PrepareStmt:            true, // 执行任一 SQL 语句时, 都会创建 Prepare Statement 并缓存, 以提高后续执行效率
		SkipDefaultTransaction: true, // 禁止在事务中进行写入操作, 性能提升约 30%
		// 覆盖默认命名策略
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false, // 表名映射不加复数, 仅仅是驼峰转为蛇形
		},
		Logger: myDBLogger, // 日志控制
	}

	// 建立 MySQL 连接
	db, err := gorm.Open(mysql.Open(dataSourceName), gormConfig) // 生成一个 *gorm.DB
	if err != nil {
		panic(fmt.Errorf("go-postery ConnectToDB : 连接到数据库出错 %s", err))
	}

	// 设置连接池相关配置
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)  // 连接池中空闲连接数目上限, 超出此上限就把相应的连接关闭掉
	sqlDB.SetMaxOpenConns(100) // 最多打开链接数目上限
	// 单个连接的连接时间上限, 超时会自动关闭, 因为数据库本身可能也对NoActive连接设置了超时时间, 我们的应对办法: 定期 ping, 或者 SetConnMaxLifetime
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 赋给全局变量 GoPosteryDB
	GoPosteryDB = db
}
