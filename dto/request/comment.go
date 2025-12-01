package request

type CreateCommentRequest struct {
	PostId   int    `json:"post_id" form:"post_id" binding:"required"`
	ParentId int    `json:"parent_id" form:"parent_id"`
	Content  string `json:"content" form:"content" binding:"required,gte=1"`
}
