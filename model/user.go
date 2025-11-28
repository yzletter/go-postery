package model

// User 定义数据库中 user 表的模型映射
type User struct {
	Id       int    `gorm:"primaryKey"`      // 用户 ID
	Name     string `gorm:"column:name"`     // 用户名
	PassWord string `gorm:"column:password"` // 用户密码 MD5 后的结果
}
