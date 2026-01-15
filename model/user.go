package model

import "time"

// User 定义数据库模型
type User struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement:false"` // 用户 ID
	Status    int        `gorm:"column:status"`                            // 状态 1 正常, 2 封禁, 3 注销
	CreatedAt time.Time  `gorm:"column:created_at"`                        // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"`                        // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"`                        // 逻辑删除时间
}

func (u User) TableName() string {
	return "users"
}

// UserIdentity 用户登录凭证
type UserIdentity struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement:false"` // 记录 ID
	UserID     int64      `gorm:"column:user_id"`                           // 用户 ID
	AuthType   int        `gorm:"column:auth_type"`                         // 校验方式
	Identifier string     `gorm:"column:identifier"`                        // 凭证
	IsVerified int        `gorm:"column:is_verified"`                       // 是否已验证
	VerifiedAt *time.Time `gorm:"column:verified_at"`                       // 验证时间
	CreatedAt  time.Time  `gorm:"column:created_at"`                        // 创建时间
	UpdatedAt  time.Time  `gorm:"column:updated_at"`                        // 更新时间
}

func (u UserIdentity) TableName() string {
	return "user_identities"
}

// UserPassword 用户密码
type UserPassword struct {
	UserID       int64     `gorm:"column:user_id;primaryKey;autoIncrement:false"` // 用户 ID
	PasswordHash string    `gorm:"column:password_hash"`                          // 用户密码
	CreatedAt    time.Time `gorm:"column:created_at"`                             // 创建时间
	UpdatedAt    time.Time `gorm:"column:updated_at"`                             // 更新时间
}

func (u UserPassword) TableName() string {
	return "user_passwords"
}

// UserProfile 用户个人资料
type UserProfile struct {
	UserID      int64      `gorm:"column:user_id;primaryKey;autoIncrement:false"` // 用户 ID
	NickName    string     `gorm:"column:nickname"`                               // 用户昵称
	Avatar      *string    `gorm:"column:avatar"`                                 // 头像 URL
	Bio         *string    `gorm:"column:bio"`                                    // 个性签名
	Gender      int        `gorm:"column:gender"`                                 // 性别 0 空, 1 男, 2 女, 3 其他
	BirthDay    *time.Time `gorm:"column:birthday"`                               // 生日
	Location    *string    `gorm:"column:location"`                               // 地区
	Country     *string    `gorm:"column:country"`                                // 国家
	LastLoginIP *string    `gorm:"column:last_login_ip"`                          // 最后登录 IP
	LastLoginAt *time.Time `gorm:"column:last_login_at"`                          // 最后登录时间
	CreatedAt   time.Time  `gorm:"column:created_at"`                             // 创建时间
	UpdatedAt   time.Time  `gorm:"column:updated_at"`                             // 更新时间
	DeletedAt   *time.Time `gorm:"column:deleted_at"`                             // 逻辑删除时间
}

func (u UserProfile) TableName() string {
	return "user_profiles"
}
