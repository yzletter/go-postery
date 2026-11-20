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
// 86400 / 432 = 200
// 两天同一时刻创建的文章 A 和 B
// A 只需要两百票就能与 B 的热度相同, 也就是只需要 200 个 432 就能续一天热度
const (
	// Post 相关排名系数

	LikeCoefficient    = 432 / 3
	CommentCoefficient = 432 / 2

	// User 相关排名系数

	FollowCoefficient = 432
)
