package model

import "time"

// User 用户最小单位
type User struct {
	ID        int64      `gorm:"column:id;primaryKey;autoIncrement:false"` // 用户 ID
	Status    int        `gorm:"column:status"`                            // 状态 1 正常, 2 封禁, 3 注销
	Role      int        `gorm:"column:role"`                              // 用户权限 0 普通 1 管理员
	CreatedAt time.Time  `gorm:"column:created_at"`                        // 创建时间
	UpdatedAt time.Time  `gorm:"column:updated_at"`                        // 更新时间
	DeletedAt *time.Time `gorm:"column:deleted_at"`                        // 逻辑删除时间
}

func (u User) TableName() string {
	return "users"
}

// UserProfile 用户个人资料
type UserProfile struct {
	UserID         int64      `gorm:"column:user_id;primaryKey;autoIncrement:false"` // 用户 ID
	Nickname       string     `gorm:"column:nickname"`                               // 用户昵称
	NicknameActive *string    `gorm:"column:nickname_active;->"`                     // 昵称是否有效
	Avatar         *string    `gorm:"column:avatar"`                                 // 头像 URL
	Bio            *string    `gorm:"column:bio"`                                    // 个性签名
	Gender         int        `gorm:"column:gender"`                                 // 性别 0 空, 1 男, 2 女, 3 其他
	Birthday       *time.Time `gorm:"column:birthday"`                               // 生日
	Location       *string    `gorm:"column:location"`                               // 地区
	Country        *string    `gorm:"column:country"`                                // 国家
	LastLoginIP    *string    `gorm:"column:last_login_ip"`                          // 最后登录 IP
	LastLoginAt    *time.Time `gorm:"column:last_login_at"`                          // 最后登录时间
	CreatedAt      time.Time  `gorm:"column:created_at"`                             // 创建时间
	UpdatedAt      time.Time  `gorm:"column:updated_at"`                             // 更新时间
	DeletedAt      *time.Time `gorm:"column:deleted_at"`                             // 逻辑删除时间
}

func (u UserProfile) TableName() string {
	return "user_profiles"
}
