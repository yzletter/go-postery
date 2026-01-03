package order

type PayRequest struct {
	UserID int64 `json:"user_id"`
	GiftID int64 `json:"gift_id"`
}

type GiveUpRequest struct {
	UserID int64 `json:"user_id"`
	GiftID int64 `json:"gift_id"`
}
