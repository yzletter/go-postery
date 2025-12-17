package comment

type CreateRequest struct {
	PostID   int64  `json:"post_id,string" binding:"required"`
	ParentID int64  `json:"parent_id,string"`
	ReplyID  int64  `json:"reply_id,string"`
	Content  string `json:"content"  binding:"required,gte=1"`
}
