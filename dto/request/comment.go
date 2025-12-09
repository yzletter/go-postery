package request

type CreateCommentRequest struct {
	PostId   int    `json:"post_id,string"  binding:"required"`
	ParentId int    `json:"parent_id,string"`
	ReplyId  int    `json:"reply_id,string"`
	Content  string `json:"content"  binding:"required,gte=1"`
}
