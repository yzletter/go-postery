package repository

import (
	"context"

	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository/cache"
	"github.com/yzletter/go-postery/repository/dao"
)

type tagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
}

func NewTagRepository(tagDAO dao.TagDAO, tagCache cache.TagCache) TagRepository {
	return &tagRepository{dao: tagDAO, cache: tagCache}
}

func (repo *tagRepository) Create(ctx context.Context, tag *model.Tag) error {
	err := repo.dao.Create(ctx, tag)
	if err != nil {
		return err
	}
	return nil
}

func (repo *tagRepository) GetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	tag, err := repo.dao.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (repo *tagRepository) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	tag, err := repo.dao.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (repo *tagRepository) Bind(ctx context.Context, postTag *model.PostTag) error {
	err := repo.dao.Bind(ctx, postTag)
	if err != nil {
		return err
	}
	return nil
}

func (repo *tagRepository) DeleteBind(ctx context.Context, pid, tid int64) error {
	err := repo.dao.DeleteBind(ctx, pid, tid)
	if err != nil {
		return err
	}
	return nil
}

func (repo *tagRepository) FindTagsByPostID(ctx context.Context, pid int64) ([]string, error) {
	tags, err := repo.dao.FindTagsByPostID(ctx, pid)
	if err != nil {
		return nil, err
	}

	return tags, nil
}
