package infraMySQL

import "time"

type PoolConfig struct {
	MaxIdleConns    int           // 空闲连接数目上限
	MaxOpenConns    int           // 最多打开链接数目上限
	ConnMaxLifetime time.Duration // 单个连接的连接时间上限
}
