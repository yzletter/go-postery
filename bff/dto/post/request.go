package post

type CreatePostRequest struct {
	Title       string   `json:"title"   binding:"required,gte=1"`  // 长度>=1
	Content     string   `json:"content"  binding:"required,gte=1"` // 长度>=1
	ContentType int      `json:"content_type" binding:"required"`
	Tags        []string `json:"tags"`
}
type UpdateRequest struct {
	Title   string   `json:"title"  binding:"required,gte=1"`    // 长度>=1
	Content string   `json:"content"   binding:"required,gte=1"` // 长度>=1
	Tags    []string `json:"tags"`
}

type CreateCommentRequest struct {
	ParentID int64  `json:"parent_id,string"`
	ReplyID  int64  `json:"reply_id,string"`
	Content  string `json:"content"  binding:"required,gte=1"`
}
