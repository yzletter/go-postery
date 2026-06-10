package model

import "time"

const (
	CodeStatusSent = iota
	CodeStatusVerified
)

// VerificationCode 验证码发送记录，用于审计和后续风控统计。
type VerificationCode struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	Biz        int        `gorm:"column:biz"`
	Identifier string     `gorm:"column:identifier"`
	CodeHash   string     `gorm:"column:code_hash"`
	Status     int        `gorm:"column:status"`
	ExpiresAt  time.Time  `gorm:"column:expires_at"`
	VerifiedAt *time.Time `gorm:"column:verified_at"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at"`
}

func (v VerificationCode) TableName() string {
	return "verification_codes"
}
