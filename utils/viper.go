package utils

import (
	"fmt"
	"path"

	"github.com/spf13/viper"
)

const (
	YAML = "yaml"
)

// InitViper 初始化 Viper 读取配置, 传入目录、文件名、文件类型, 返回一个 viper.Viper 指针
func InitViper(dir, fileName, fileType string) *viper.Viper {
	config := viper.New()          // 创建 *Viper 对象
	config.AddConfigPath(dir)      // 设置配置文件所在目录
	config.SetConfigName(fileName) // 设置配置文件名 (无路径, 无后缀)
	config.SetConfigType(fileType) // 设置文件类型

	// 尝试解析配置文件
	if err := config.ReadInConfig(); err != nil {
		configFile := path.Join(dir, fileName) + "." + fileType // 完整配置文件路径
		// 系统初始化过程中发生错误直接 panic, logger 还未初始化, 不能用 logger.fatal()
		panic(fmt.Errorf("go-postery InitViper : 解析 [%s] 配置文件出错 %s", configFile, err))
	}

	return config
}
