package database

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-sql-driver/mysql"
	"github.com/yzletter/go-postery/database/model"
)

// RegisterUser 传入 name 和 password 注册新用户, 返回 Id 和可能的错误
func RegisterUser(name, password string) (int, error) {
	// 将模型绑定到结构体
	user := model.User{
		Name:     name,
		PassWord: password,
	}

	// 到 MySQL 中创建新记录
	err := GoPosteryDB.Create(&user).Error // 需要传指针
	// 错误处理
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) { // 判断是否为 MySQL 错误
			if mysqlErr.Number == 1062 { // Unique Key 冲突
				return 0, fmt.Errorf("用户[%s]已存在", name)
			}
		}
		// 记录日志, 方便后续人工定位问题所在
		slog.Error("go-postery RegisterUser : 用户注册失败", "name", name, "error", err)
		return 0, fmt.Errorf("用户注册失败，请稍后重试")
	}

	// 返回 Id
	return user.Id, nil
}
