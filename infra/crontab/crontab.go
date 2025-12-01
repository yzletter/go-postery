package crontab

import (
	"log/slog"

	"github.com/robfig/cron/v3"
	infraMySQL "github.com/yzletter/go-postery/infra/mysql"
	infraRedis "github.com/yzletter/go-postery/infra/redis"
)

func InitCrontab() {
	crontab := cron.New()

	_, err := crontab.AddFunc("*/10 * * * *", infraMySQL.Ping) // 分别代表 分 时 周 月 星期, 每十分钟 ping 一次 MySQL
	if err != nil {
		slog.Error("crontab add func failed", "error", err)
	}

	_, err = crontab.AddFunc("*/10 * * * *", infraRedis.Ping) // 分别代表 分 时 周 月 星期, 每十分钟 ping 一次 Redis
	if err != nil {
		slog.Error("crontab add func failed", "error", err)
	}

	crontab.Start()
}
