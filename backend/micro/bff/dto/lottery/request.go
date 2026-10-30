package lottery

type PayRequest struct {
	UserID      int64 `json:"user_id,string"`
	GiftID      int64 `json:"gift_id,string"`
	TempOrderID int64 `json:"temp_order_id,string"`
}

type GiveUpRequest struct {
	UserID      int64 `json:"user_id,string"`
	GiftID      int64 `json:"gift_id,string"`
	TempOrderID int64 `json:"temp_order_id,string"`
}
