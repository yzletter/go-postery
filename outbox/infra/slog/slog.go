package slog

import (
	"fmt"
	"log/slog"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// InitSlog 初始化 Slog
func InitSlog(logFileName string) {
	// 设置 rotatelogs 滚动日志相关配置
	logFile, err := rotatelogs.New(
		logFileName+".%Y%m%d%H",                  // 日志文件路径
		rotatelogs.WithLinkName(logFileName),     // 创建软链接指向最新的一份日志
		rotatelogs.WithRotationTime(1*time.Hour), // 设置滚动时间, 每小时滚动一次
		rotatelogs.WithMaxAge(1*time.Hour),       // 设置日志保存时间, 或使用 WithRotationCount 只保留最近的几份日志
	)
	if err != nil {
		panic(fmt.Errorf("go-postery InitSlog : 滚动日志配置出错 %s", err))
	}

	// 设置 Slog 相关配置
	slogConfig := &slog.HandlerOptions{
		AddSource: true,           // 报告文件名和行号
		Level:     slog.LevelInfo, // 设置日志最低级别
		// 用 Go 标准时间格式替换默认时间格式
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey { // 如果 Key == "time"
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05.000")) // 替换 Value
			}
			return a
		},
	}

	// 构造 logger
	slogHandler := slog.NewTextHandler( // JSON 格式
		logFile,    // 指定文件
		slogConfig, // 相关配置
	)
	logger := slog.New(slogHandler)

	// 设置为全局 logger
	slog.SetDefault(logger)
}
