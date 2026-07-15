package model

// InterviewQuestion Milvus 数据模型
type InterviewQuestion struct {
	ID         int64     `json:"id" milvus:"name:id"` // 题目 ID
	Content    string    `json:"content" milvus:"name:content"`
	Vector     []float32 `json:"vector" milvus:"name:vector"`
	Metadata   []byte    `json:"metadata" milvus:"name:metadata"`
	UserID     int64     `json:"user_id" milvus:"name:user_id"` // 用户 ID
	SourceFile string    `json:"source_file" milvus:"name:source_file"`
}

type Question struct {
	ID         int64    `json:"id"` // 题目 ID
	Content    string   `json:"content"`
	Type       string   `json:"type"`
	Difficulty string   `json:"difficulty"`
	Skills     []string `json:"skills"`
	FollowUps  []string `json:"follow_ups"`
	Reference  string   `json:"reference"`
}
