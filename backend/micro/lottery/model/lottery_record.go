package model

type LotteryRecord struct {
	RequestID int64 // 预占请求 ID
	UserID    int64 // 用户 ID
	GiftID    int64 // 奖品 ID
}
