package model

type TempOrder struct {
	ID     int64 `json:"id,string"`
	UserID int64 `json:"user_id,string"`
	GiftID int64 `json:"gift_id,string"`
}

type LotteryResult struct {
	Gift        *Gift
	TempOrderID int64
}
