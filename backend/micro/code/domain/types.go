package domain

import "time"

type BizType int

const (
	BizSMS BizType = iota + 1
	BizEmail
)

type CodeRecord struct {
	ID         int64
	Biz        BizType
	Identifier string
	CodeHash   string
	ExpiresAt  time.Time
}
