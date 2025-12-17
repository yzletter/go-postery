package post

type CreateRequest struct {
	ID      int64    `json:"id,string" `
	Title   string   `json:"title"   binding:"required,gte=1"`  // 长度>=1
	Content string   `json:"content"  binding:"required,gte=1"` // 长度>=1
	Tags    []string `json:"tags"`
}
type UpdateRequest struct {
	ID      int64    `json:"id,string" `
	Title   string   `json:"title"  binding:"required,gte=1"`    // 长度>=1
	Content string   `json:"content"   binding:"required,gte=1"` // 长度>=1
	Tags    []string `json:"tags"`
}
