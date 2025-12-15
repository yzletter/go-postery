package repository

import (
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
