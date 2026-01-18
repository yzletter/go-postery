package model

import "time"

// AuthIdentity 用户登录凭证
type AuthIdentity struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement:false"` // 记录 ID
	UserID     int64      `gorm:"column:user_id"`                           // 用户 ID
	AuthType   int        `gorm:"column:auth_type"`                         // 校验方式
	Identifier string     `gorm:"column:identifier"`                        // 凭证
	IsVerified int        `gorm:"column:is_verified"`                       // 是否已验证
	VerifiedAt *time.Time `gorm:"column:verified_at"`                       // 验证时间
	CreatedAt  time.Time  `gorm:"column:created_at"`                        // 创建时间
	UpdatedAt  time.Time  `gorm:"column:updated_at"`                        // 更新时间
}

func (u AuthIdentity) TableName() string {
	return "auth_identities"
}

// AuthPassword 用户密码
type AuthPassword struct {
	UserID       int64     `gorm:"column:user_id;primaryKey;autoIncrement:false"` // 用户 ID
	PasswordHash string    `gorm:"column:password_hash"`                          // 用户密码
	CreatedAt    time.Time `gorm:"column:created_at"`                             // 创建时间
	UpdatedAt    time.Time `gorm:"column:updated_at"`                             // 更新时间
}

func (u AuthPassword) TableName() string {
	return "auth_passwords"
}

func AuthTypeFromBiz(biz CodeBiz) int {
	switch biz {
	case SMSCode:
		return 1
	case EmailCode:
		return 2
	default:
		return 0
	}
}

type AuthAggregate struct {
	User         *User
	UserProfile  *UserProfile
	AuthPassword *AuthPassword
	AuthIdentity *AuthIdentity
	Events       []*Event
}
