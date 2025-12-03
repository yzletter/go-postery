package request

type CreateCommentRequest struct {
	PostId   int    `json:"post_id"  binding:"required"`
	ParentId int    `json:"parent_id"`
	Content  string `json:"content"  binding:"required,gte=1"`
}
