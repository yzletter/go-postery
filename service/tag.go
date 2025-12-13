package service

import (
	"log/slog"

	tagRepository "github.com/yzletter/go-postery/repository/tag"
	"github.com/yzletter/go-postery/utils"
)

type TagService struct {
	TagCacheRepo *tagRepository.TagCacheRepository
	TagDBRepo    *tagRepository.TagDBRepository
}

func NewTagService(tagDBRepo *tagRepository.TagDBRepository, tagCacheRepo *tagRepository.TagCacheRepository) *TagService {
	return &TagService{
		TagCacheRepo: tagCacheRepo,
		TagDBRepo:    tagDBRepo,
	}
}

func (svc *TagService) Create(name string) (int, error) {
	// 获得唯一标识符
	slug := utils.Slugify(name)
	tid, err := svc.TagDBRepo.Create(name, slug)
	if err != nil {
		return 0, err
	}
	return tid, nil
}

// Bind 将 Tags 绑定到 post
func (svc *TagService) Bind(pid int, tags []string) {
	slog.Info("tags", "tags", tags)
	for _, tag := range tags {
		tid, err := svc.TagDBRepo.Exist(tag)
		if err != nil {
			// tag 不存在需要创建
			tid, err = svc.Create(tag)
			if err != nil {
				// 虽然有错误, 但尽可能的多绑定
				slog.Error("Tag Not Exist And Created Failed", "error", err)
				continue
			}
		}

		err = svc.TagDBRepo.Bind(pid, tid)
		if err != nil {
			slog.Error("Tag 绑定失败", "error", err)
		}
	}
}
func (svc *TagService) FindTagsByPostID(pid int) []string {
	res, _ := svc.TagDBRepo.FindTagsByPostID(pid)
	return res
}
