package repository

import (
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type commentRepository struct {
	dao   dao.CommentDAO
	cache cache.CommentCache
}

func NewCommentRepository(commentDAO dao.CommentDAO, commentCache cache.CommentCache) CommentRepository {
	return &commentRepository{dao: commentDAO, cache: commentCache}
}
