package qdrant

import (
	"log/slog"
	"sync"

	"github.com/qdrant/go-client/qdrant"
	"github.com/yzletter/go-postery/infra/viper"
)

var (
	client *qdrant.Client
	once   sync.Once
)

func Init(confDir, confFileName, confFileType string) *qdrant.Client {
	once.Do(func() {
		// 读取 Qdrant 相关配置
		vip := viper.InitViper(confDir, confFileName, confFileType) // 初始化一个 Viper 进行配置读取
		host := vip.GetString("qdrant.host")
		port := vip.GetInt("qdrant.port")

		var err error
		client, err = qdrant.NewClient(&qdrant.Config{
			Host: host,
			Port: port,
		})
		if err != nil {
			slog.Error("初始化 Qdrant 失败 ...", "error", err)
		}
	})

	return client
}

func Close() {
	if client != nil {
		err := client.Close()
		if err != nil {
			slog.Error("关闭 Qdrant 失败 ...")
			return
		}
		slog.Error("关闭 Qdrant 成功 ...")
	}
}
