package service

import (
	"log/slog"

	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/utils"
)

type tagService struct {
	TagRepo repository.TagRepository
}

func NewTagService(tagRepo repository.TagRepository) TagService {
	return &tagService{
		TagRepo: tagRepo,
	}
}

func (svc *tagService) Create(name string) (int, error) {
	// 获得唯一标识符
	tagName := name
	slug := utils.Slugify(name)
	tid, err := svc.TagRepo.Create(tagName, slug)
	if err != nil {
		return 0, err
	}
	return tid, nil
}

// Bind 将 Tags 绑定到 post
func (svc *tagService) Bind(pid int, tags []string) {
	slog.Info("tags", "tags", tags)
	for _, tag := range tags {
		tid, err := svc.TagRepo.Exist(tag)
		if err != nil {
			// tag 不存在需要创建
			tid, err = svc.Create(tag)
			if err != nil {
				// 虽然有错误, 但尽可能的多绑定
				slog.Error("Tag Not Exists And CreatedAt Failed", "error", err)
				continue
			}
		}

		err = svc.TagRepo.Bind(pid, tid)
		if err != nil {
			slog.Error("Tag 绑定失败", "error", err)
		}
	}
}
func (svc *tagService) FindTagsByPostID(pid int) []string {
	res, _ := svc.TagRepo.FindTagsByPostID(pid)
	return res
}
