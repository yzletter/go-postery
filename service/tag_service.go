package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
	"github.com/yzletter/go-postery/utils"
)

type tagService struct {
	tagRepo repository.TagRepository
	idGen   ports.IDGenerator
}

func NewTagService(tagRepo repository.TagRepository, idGen ports.IDGenerator) TagService {
	return &tagService{
		tagRepo: tagRepo,
		idGen:   idGen,
	}
}

// Create 新建 Tag
func (svc *tagService) Create(ctx context.Context, name string) (int64, error) {
	// 获得唯一标识符
	tagName := name
	slug := utils.Slugify(name)

	tag := &model.Tag{
		ID:   svc.idGen.NextID(),
		Name: tagName,
		Slug: slug,
	}

	err := svc.tagRepo.Create(ctx, tag)
	if err != nil {
		if !errors.Is(err, repository.ErrUniqueKey) {
			return 0, errno.ErrServerInternal
		}
	}
	return tag.ID, nil
}

// Bind 将 Tags 绑定到 post
func (svc *tagService) Bind(ctx context.Context, pid int64, tags []string) error {
	for _, tagName := range tags {
		// todo GetOrCreateByName 并发
		tag, err := svc.tagRepo.GetByName(ctx, tagName) // 查 tid
		if err != nil {
			newTag := &model.Tag{
				ID:   svc.idGen.NextID(),
				Name: tagName,
				Slug: utils.Slugify(tagName),
			}
			err = svc.tagRepo.Create(ctx, newTag) // 没有 tag 就新建
			if err != nil {
				continue
			}
			tag = newTag
		}

		// 绑定
		postTag := &model.PostTag{
			ID:     svc.idGen.NextID(),
			PostID: pid,
			TagID:  tag.ID,
		}
		if err = svc.tagRepo.Bind(ctx, postTag); err != nil {
			slog.Error("Bind Post Tag Failed", "error", err)
			if errors.Is(err, repository.ErrUniqueKey) {
				return errno.ErrTagDuplicatedBind
			}
			return errno.ErrServerInternal
		}
	}

	return nil
}

// FindTagsByPostID 根据帖子 ID 查找 Tag
func (svc *tagService) FindTagsByPostID(ctx context.Context, pid int64) ([]string, error) {
	var empty []string
	res, err := svc.tagRepo.FindTagsByPostID(ctx, pid)
	if err != nil {
		return empty, errno.ErrServerInternal
	}
	return res, nil
}
