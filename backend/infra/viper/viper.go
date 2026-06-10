package viper

import (
	"fmt"
	"path"

	"github.com/spf13/viper"
)

const YAML = "yaml"

func InitViper(dir, fileName, fileType string) *viper.Viper {
	config := viper.New()
	config.AddConfigPath(dir)
	config.SetConfigName(fileName)
	config.SetConfigType(fileType)

	if err := config.ReadInConfig(); err != nil {
		configFile := path.Join(dir, fileName) + "." + fileType
		panic(fmt.Errorf("go-postery InitViper : parse [%s] failed: %w", configFile, err))
	}

	return config
}
