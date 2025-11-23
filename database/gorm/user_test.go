package database_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"

	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/utils"
)

// Hash 返回字符串 MD5 哈希后 32 位的十六进制编码结果
func hash(password string) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	digest := hasher.Sum(nil)
	return hex.EncodeToString(digest)
}

func init() {
	utils.InitSlog("../../log/go_postery.log")
	database.ConnectToDB("../../conf", "db", "yaml", "../../log")
}

func TestRegisterUser(t *testing.T) {
	// 注册一次 yzletter, 结果应为成功
	id1, err := database.RegisterUser("yzletter", hash("123456"))
	if err != nil {
		fmt.Printf("用户[%d]注册失败 \n", id1)
		t.Fatal()
	} else {
		fmt.Printf("用户[%d]注册成功 \n", id1)
	}

	// 再注册一次 yzletter, 结果应为失败
	id2, err := database.RegisterUser("yzletter", hash("123456"))
	if err == nil {
		fmt.Printf("用户[%d]重复成功 \n", id2)
		t.Fatal()
	} else {
		fmt.Println("用户重复注册")
	}
}

// go test -v ./database/gorm -run=^TestRegisterUser$ -count=1
