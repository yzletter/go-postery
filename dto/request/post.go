package request

type CreatePostRequest struct {
	Id      int      `json:"id,string" form:"id"`
	Title   string   `json:"title" form:"title"  binding:"required,gte=1"`     // 长度>=1
	Content string   `json:"content" form:"content"  binding:"required,gte=1"` // 长度>=1
	Tags    []string `json:"tags"`
}
type UpdatePostRequest struct {
	Id      int    `json:"id,string" form:"id"`
	Title   string `json:"title" form:"title"  binding:"required,gte=1"`     // 长度>=1
	Content string `json:"content" form:"content"  binding:"required,gte=1"` // 长度>=1
}
