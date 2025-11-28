package database_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"

	database "github.com/yzletter/go-postery/database/gorm"
	"github.com/yzletter/go-postery/utils"
)

func init() {
	utils.InitSlog("../../log/go_postery.log")
	database.ConnectToMySQL("../../conf", "db", utils.YAML, "../../log")
}

func TestRegisterUser(t *testing.T) {
	// 注册一次 yzletter, 结果应为成功
	id1, err := database.RegisterUser("yzletter2", hash("123456"))
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

func TestLogOffUser(t *testing.T) {
	var uid = 7
	// 删前查询
	user := database.GetUserById(uid)
	fmt.Println(user)

	err := database.LogOffUser(uid)
	if err != nil {
		t.Fatal(err)
	} else {
		fmt.Println("首次删除成功")
	}

	// 删完后再查询
	user = database.GetUserById(uid)
	fmt.Println(user)

	// 再删一次
	err = database.LogOffUser(uid)
	if err == nil {
		t.Fatal(err)
	} else {
		fmt.Println("重复删除失败")
	}
}

func TestUpdatePassword(t *testing.T) {
	var uid = 9
	err := database.UpdatePassword(uid, hash("123456"), hash("654321"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUserByName(t *testing.T) {
	// 注册一次 yzletter, 结果应为成功
	id, _ := database.RegisterUser("getuserbyname", hash("123456"))
	user := database.GetUserById(id)
	if user != nil {
		result := database.GetUserByName(user.Name)
		fmt.Println(result)
	}
	_ = database.LogOffUser(id)
}

// 返回字符串 MD5 哈希后 32 位的十六进制编码结果
func hash(password string) string {
	hasher := md5.New()
	hasher.Write([]byte(password))
	digest := hasher.Sum(nil)
	return hex.EncodeToString(digest)
}

// go test -v ./database/gorm -run=^TestRegisterUser$ -count=1
// go test -v ./database/gorm -run=^TestLogOffUser$ -count=1
// go test -v ./database/gorm -run=^TestUpdatePassword$ -count=1
// go test -v ./database/gorm -run=^TestGetUserByName$ -count=1
