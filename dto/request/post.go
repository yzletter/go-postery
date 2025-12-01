package request

type CreateRequest struct {
	Id      int    `json:"id" form:"id"`
	Title   string `json:"title" form:"title"  binding:"required,gte=1"`     // 长度>=1
	Content string `json:"content" form:"content"  binding:"required,gte=1"` // 长度>=1
}
type UpdateRequest struct {
	Id      int    `json:"id" form:"id"`
	Title   string `json:"title" form:"title"  binding:"required,gte=1"`     // 长度>=1
	Content string `json:"content" form:"content"  binding:"required,gte=1"` // 长度>=1
}
