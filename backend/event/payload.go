package event

import "time"

// NewUserEventPayload 新用户创建
type NewUserEventPayload struct {
	ID int64 `json:"id,string"`
}

// NewPostEventPayload 新帖子创建
type NewPostEventPayload struct {
	ID int64 `json:"id,string"`
}

// NewReadEventPayload 新阅读相关互动
type NewReadEventPayload struct {
	ID      int64
	UserID  int64
	PostID  int64
	EventAt time.Time
}

// NewFollowEventPayload 新关注相关互动
type NewFollowEventPayload struct {
	ID         int64
	Follower   int64
	Followee   int64
	FollowType int
	EventAt    time.Time
}

const (
	Follow = iota + 1
	Unfollow
)

// NewLikeEventPayload 新点赞相关互动
type NewLikeEventPayload struct {
	ID       int64
	UserID   int64
	PostID   int64
	LikeType LikeType
	EventAt  time.Time
}

type LikeType int

const (
	Like LikeType = iota + 1
	Unlike
)

// NewCommentEventPayload 新相关互动评论相关互动
type NewCommentEventPayload struct {
	ID          int64
	UserID      int64
	PostID      int64
	Cnt         int
	CommentType CommentType
	EventAt     time.Time
}

type CommentType int

const (
	Create CommentType = iota + 1
	Del
)

// UpdateScoreEventPayload 更新分数
type UpdateScoreEventPayload struct {
	ID    int64
	Biz   UpdateType
	BizID int64
}

type UpdateType int

const (
	UpdateUserScore UpdateType = iota + 1
	UpdatePostScore
)
