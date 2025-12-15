package model

import "time"

// User 定义数据库模型
type User struct {
	ID           int64      `gorm:"primaryKey"`                    // 用户 ID
	Username     string     `gorm:"column:username"`               // 用户名
	Email        string     `gorm:"column:email"`                  // 邮箱
	PasswordHash string     `json:"-" gorm:"column:password_hash"` // 密码哈希
	Avatar       string     `gorm:"column:avatar"`                 // 头像 URL
	Bio          string     `gorm:"column:bio"`                    // 个性签名
	Gender       int        `gorm:"column:gender"`                 // 性别 0 空, 1 男, 2 女, 3 其他
	BirthDay     *time.Time `gorm:"column:birthday"`               // 生日
	Location     string     `gorm:"column:location"`               // 地区
	Country      string     `gorm:"column:country"`                // 国家
	Status       int        `gorm:"column:status"`                 // 状态 1 正常, 2 封禁, 3 注销
	LastLoginIP  string     `gorm:"column:last_login_ip"`          // 最后登录 IP
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`          // 最后登录时间
	CreatedAt    time.Time  `gorm:"column:created_at"`             // 创建时间
	UpdatedAt    time.Time  `gorm:"column:updated_at"`             // 更新时间
	DeletedAt    *time.Time `gorm:"column:deleted_at"`             // 逻辑删除时间
}

// TableName 指定表名
func (u User) TableName() string {
	return "users"
}
