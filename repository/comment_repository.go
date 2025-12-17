package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
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

func (repo *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	err := repo.dao.Create(ctx, comment)
	if err != nil {
		return toRepositoryErr(err)
	}

	return nil
}

func (repo *commentRepository) GetByID(ctx context.Context, id int64) (*model.Comment, error) {
	c, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	return c, nil
}

func (repo *commentRepository) Delete(ctx context.Context, id int64) (int, error) {
	cnt, err := repo.dao.Delete(ctx, id)
	if err != nil {
		return cnt, toRepositoryErr(err)
	}

	return cnt, nil
}

func (repo *commentRepository) GetByPostID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error) {
	total, comments, err := repo.dao.GetByPostID(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, comments, nil
}

func (repo *commentRepository) GetRepliesByParentIDs(ctx context.Context, ids []int64) ([]*model.Comment, error) {
	comments, err := repo.dao.GetRepliesByParentIDs(ctx, ids)
	if err != nil {
		return nil, toRepositoryErr(err)
	}

	return comments, nil
}
