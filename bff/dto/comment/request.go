package comment

type CreateRequest struct {
	ParentID int64  `json:"parent_id,string"`
	ReplyID  int64  `json:"reply_id,string"`
	Content  string `json:"content"  binding:"required,gte=1"`
}
