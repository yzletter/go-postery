package model

import "time"

// User 定义数据库中 user 表的模型映射
type User struct {
	Id          int       `gorm:"primaryKey"`           // ID 雪花算法
	Name        string    `gorm:"column:name"`          // 用户名
	PassWord    string    `gorm:"column:password"`      // 密码 MD5 后的结果
	Email       string    `gorm:"column:email"`         // 邮箱
	Avatar      string    `gorm:"column:avatar"`        // 头像 URL
	Bio         string    `gorm:"column:bio"`           // 个性签名
	Gender      int       `gorm:"column:gender"`        // 性别: 0 表示空, 1 表示男, 2 表示女, 3 表示其它
	BirthDay    time.Time `gorm:"column:birthday"`      // 生日
	Location    string    `gorm:"column:location"`      // 地区
	Country     string    `gorm:"column:country"`       // 国家
	Status      int       `gorm:"column:status"`        // 状态: 0 表示空, 1 表示正常, 2 表示封禁, 3 表示注销
	LastLoginIP string    `gorm:"column:last_login_ip"` // 最近一次登录 IP
}
