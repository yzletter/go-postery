package domain

type User struct {
	ID    int64
	Score int64
}

type Post struct {
	ID    int64
	Score int64
}

// 业务区分 Biz
const (
	// BizUser 用户业务 Biz
	BizUser = iota + 1
	// BizPost 帖子业务 Biz
	BizPost
)

// 排名计算参数
const (
	LikeCoefficient    = 100
	CommentCoefficient = 100
	FollowCoefficient  = 100
)
