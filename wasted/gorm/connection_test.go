package database_test

//
//import (
//	"testing"
//
//	"github.com/yzletter/go-postery/repository/gorm"
//	"github.com/yzletter/go-postery/utils"
//)
//
//func TestConnection(t *testing.T) {
//	database.ConnectToMySQL("../../conf", "db", utils.YAML, "../../log")
//	sqlDB, err := database.GoPosteryMySQLDB.DB()
//	if err != nil {
//		t.Fatalf("获取 sql.DB 失败: %v", err)
//	}
//	err = sqlDB.Ping()
//	if err != nil {
//		t.Fatalf("Ping 失败: %v", err)
//	}
//}
//
//// go test -v ./database/gorm -run=^TestConnection$ -count=1
