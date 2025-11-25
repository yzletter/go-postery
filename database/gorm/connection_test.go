package database_test

import (
	"testing"

	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/utils"
)

func TestConnection(t *testing.T) {
	database.ConnectToDB("../../conf", "db", utils.YAML, "../../log")
	sqlDB, err := database.GoPosteryDB.DB()
	if err != nil {
		t.Fatalf("获取 sql.DB 失败: %v", err)
	}
	err = sqlDB.Ping()
	if err != nil {
		t.Fatalf("Ping 失败: %v", err)
	}
}

// go test -v ./database/gorm -run=^TestConnection$ -count=1
