package service

import repository "github.com/yzletter/go-postery/repository/tag"

type TagService struct {
	TagCacheRepo *repository.TagCacheRepository
	TagDBRepo    *repository.TagDBRepository
}

func NewTagService(tagCacheRepo *repository.TagCacheRepository, tagDBRepo *repository.TagDBRepository) *TagService {
	return &TagService{
		TagCacheRepo: tagCacheRepo,
		TagDBRepo:    tagDBRepo,
	}
}

func (svc *TagService) Create(name string) error {
	slug := ""
	svc.TagDBRepo.Create(name, slug)
	return nil
}
