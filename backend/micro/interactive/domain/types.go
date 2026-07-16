package domain

import (
	"time"

	"github.com/yzletter/go-postery/backend/micro/interactive/model"
)

type Comment struct {
	ID        int64     // 评论 ID
	PostID    int64     // 帖子 ID
	ParentID  int64     // 父评论 ID, 可为空
	ReplyID   int64     // 回复的评论 ID, 可为空
	UserID    int64     // 评论者 ID
	Content   string    // 评论内容
	CreatedAt time.Time // 评论时间
}

type Follow struct {
	ID         int64
	FollowerID int64
	FolloweeID int64
}

type Like struct {
	ID     int64
	UserID int64
	PostID int64
}

type PostInter struct {
	ReadCnt     int64
	LikeCnt     int64
	CommentCnt  int64
	CalculateAt time.Time
}

type UserInter struct {
	FollowCnt   int64
	CalculateAt time.Time
}

type BizType int

const (
	BizRead BizType = iota + 1
	BizLike
	BizComment
	BizFollow
)

type FollowType int

const (
	FollowWrong            = iota - 1
	FollowNone  FollowType = iota
	FollowIFollow
	FollowFollowMe
	FollowMutual
)

func ToCommentDomain(comment model.Comment) Comment {
	return Comment{
		ID:        comment.ID,
		PostID:    comment.PostID,
		ParentID:  comment.ParentID,
		ReplyID:   comment.ReplyID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}
}

func ToPostInterDomain(inter model.PostInteractive) PostInter {
	res := PostInter{
		ReadCnt:    inter.ReadCnt,
		LikeCnt:    inter.LikeCnt,
		CommentCnt: inter.CommentCnt,
	}
	if inter.CalculateAt != nil {
		res.CalculateAt = *inter.CalculateAt
	}
	return res
}

func ToUserInterDomain(inter model.UserInteractive) UserInter {
	res := UserInter{
		FollowCnt: inter.FollowCnt,
	}
	if inter.CalculateAt != nil {
		res.CalculateAt = *inter.CalculateAt
	}
	return res
}
