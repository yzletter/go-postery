package post

type CreateRequest struct {
	Title   string   `json:"title"   binding:"required,gte=1"`  // 长度>=1
	Content string   `json:"content"  binding:"required,gte=1"` // 长度>=1
	Tags    []string `json:"tags"`
}
type UpdateRequest struct {
	Title   string   `json:"title"  binding:"required,gte=1"`    // 长度>=1
	Content string   `json:"content"   binding:"required,gte=1"` // 长度>=1
	Tags    []string `json:"tags"`
}
