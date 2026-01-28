package order

type PayRequest struct {
	UserID int64 `json:"user_id,string"`
	GiftID int64 `json:"gift_id,string"`
}

type GiveUpRequest struct {
	UserID int64 `json:"user_id,string"`
	GiftID int64 `json:"gift_id,string"`
}
