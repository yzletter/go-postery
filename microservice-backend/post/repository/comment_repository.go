package repository

import (
	"context"

	"github.com/yzletter/go-postery/microservice-backend/post/model"
	"github.com/yzletter/go-postery/microservice-backend/post/repository/dao"
)

type commentRepository struct {
	dao dao.CommentDAO
}

func NewCommentRepository(commentDAO dao.CommentDAO) CommentRepository {
	return &commentRepository{dao: commentDAO}
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

func (repo *commentRepository) GetRepliesByParentID(ctx context.Context, id int64, pageNo, pageSize int) (int64, []*model.Comment, error) {
	total, comments, err := repo.dao.GetRepliesByParentID(ctx, id, pageNo, pageSize)
	if err != nil {
		return 0, nil, toRepositoryErr(err)
	}

	return total, comments, nil
}
