package service

import (
	"context"
	"errors"
	"log/slog"

	postdto "github.com/yzletter/go-postery/dto/post"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/utils"
)

type postService struct {
	postRepo repository.PostRepository
	userRepo repository.UserRepository
	likeRepo repository.LikeRepository
	tagRepo  repository.TagRepository
	idGen    IDGenerator // 用于生成 ID
}

func NewPostService(
	postRepo repository.PostRepository, userRepo repository.UserRepository,
	likeRepo repository.LikeRepository, tagRepo repository.TagRepository,
	idGen IDGenerator) PostService {
	return &postService{
		postRepo: postRepo,
		userRepo: userRepo,
		likeRepo: likeRepo,
		tagRepo:  tagRepo,
		idGen:    idGen,
	}
}

// Create 新建一篇帖子
func (svc *postService) Create(ctx context.Context, uid int64, title, content string) (postdto.DetailDTO, error) {
	var empty postdto.DetailDTO

	// 先查找作者
	user, err := svc.userRepo.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 创建帖子
	post := &model.Post{
		ID:      svc.idGen.NextID(),
		UserID:  uid,
		Title:   title,
		Content: content,
	}
	err = svc.postRepo.Create(ctx, post)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 雪花 ID 的帖子不会已存在, 需要排查
			slog.Error("Create Post Failed", "error", err)
		}
		return empty, errno.ErrServerInternal
	}

	return postdto.ToDetailDTO(post, user), err
}

// GetDetailById 获取帖子详情，并选择是否增加浏览量
func (svc *postService) GetDetailById(ctx context.Context, id int64, addViewCnt bool) (postdto.DetailDTO, error) {
	// 查找帖子详情
	var empty postdto.DetailDTO
	post, err := svc.postRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrPostNotFound
		}
		return empty, errno.ErrServerInternal
	}

	// 查找作者信息
	user, err := svc.userRepo.GetByID(ctx, post.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrUserNotFound
		}
		return empty, errno.ErrServerInternal
	}

	if addViewCnt {
		// 记录 ViewCount + 1
		filed := model.PostViewCount
		err = svc.postRepo.UpdateCount(ctx, post.ID, filed, 1) // 数据库中 + 1
		if err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				slog.Error("Update View Cnt Failed", "error", err)
			}
		}
		post.ViewCount += 1
	}

	postDTO := postdto.ToDetailDTO(post, user)
	return postDTO, nil
}

// GetBriefById 根据 ID 获取帖子简要信息
func (svc *postService) GetBriefById(ctx context.Context, id int64) (postdto.BriefDTO, error) {
	var empty postdto.BriefDTO

	// 获取帖子详情
	postDetailDTO, err := svc.GetDetailById(ctx, id, false) // 选择不加浏览量
	if err != nil {
		// 这里的错误是 errno 错误, 直接返回即可
		return empty, err
	}

	return postdto.BriefDTO{
		ID:        postDetailDTO.ID,
		Title:     postDetailDTO.Title,
		CreatedAt: postDetailDTO.CreatedAt,
		Author:    postDetailDTO.Author,
	}, nil
}

// Belong 判断登录用户是否是帖子作者
func (svc *postService) Belong(ctx context.Context, pid, uid int64) bool {
	// todo 优化只查 user_id 字段
	postBriefDTO, err := svc.GetBriefById(ctx, pid)
	if err != nil || uid != postBriefDTO.Author.ID {
		return false
	}
	return true
}

// Delete 删除帖子
func (svc *postService) Delete(ctx context.Context, pid, uid int64) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(ctx, pid, uid)
	if !ok {
		return errno.ErrUnauthorized
	}

	// 删除帖子
	err := svc.postRepo.Delete(ctx, pid)
	if err != nil {
		// 如果是记录不存在, 则幂等
		if !errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrServerInternal
		}
	}
	return nil
}

// Update 更新帖子
func (svc *postService) Update(ctx context.Context, pid int64, uid int64, title, content string, tags []string) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(ctx, pid, uid)
	if !ok {
		// 无权限更新
		return errno.ErrUnauthorized
	}

	tagsBefore, err := svc.tagRepo.FindTagsByPostID(ctx, pid)
	if err != nil {
		slog.Error("Get Tags_Before Failed", "error", err)
		return errno.ErrServerInternal
	}

	tagsNow := tags

	// 将切片转为集合
	hashBefore := make(map[string]struct{})
	for _, tag := range tagsBefore {
		hashBefore[tag] = struct{}{}
	}
	hashNow := make(map[string]struct{})
	for _, tag := range tagsNow {
		hashNow[tag] = struct{}{}
	}

	for _, tagName := range tagsBefore {
		if _, ok := hashNow[tagName]; !ok { // 原来有现在没有 ——> 删除
			tag, err := svc.tagRepo.GetByName(ctx, tagName) // 查 tagName
			if err != nil {
				// 这里应该是必须有 tagName 的才对
				slog.Error("Can Not Find Tid", "error", err)
				continue
			}

			// 解绑
			if err = svc.tagRepo.DeleteBind(ctx, pid, tag.ID); err != nil {
				slog.Error("Delete Bind Failed", "error", err)
			}
		}
	}

	for _, tagName := range tagsNow {
		if _, ok := hashBefore[tagName]; !ok { // 现在有原来没有 ——> 绑定
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
			}
		}
	}

	updates := map[string]any{
		"title":   title,
		"content": content,
	}

	err = svc.postRepo.Update(ctx, pid, updates) // 更新标题和正文
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrPostNotFound
		}
		return errno.ErrServerInternal
	}
	return nil
}

// ListByPage 按页获取帖子列表
func (svc *postService) ListByPage(ctx context.Context, pageNo, pageSize int) (int, []postdto.DetailDTO) {
	var empty []postdto.DetailDTO
	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPage(ctx, pageNo, pageSize)
	if err != nil {
		return 0, empty
	}

	// todo 避免性能问题，优化 SQL
	var postDTOs []postdto.DetailDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		author, err := svc.userRepo.GetByID(ctx, post.UserID)
		if err != nil {
			slog.Warn("could not get name of user", "uid", post.UserID)
			author = &model.User{}
		}
		postDTO := postdto.ToDetailDTO(post, author)
		postDTOs = append(postDTOs, postDTO)
	}
	return int(total), postDTOs
}

// ListByUid 根据作者 ID 获取帖子简要信息列表
func (svc *postService) ListByUid(ctx context.Context, uid int64, pageNo, pageSize int) (int, []postdto.BriefDTO) {
	var empty []postdto.BriefDTO
	total, posts, err := svc.postRepo.GetByUid(ctx, uid, pageNo, pageSize)
	if err != nil {
		return 0, empty
	}

	// 查找作者信息
	author, err := svc.userRepo.GetByID(ctx, uid)
	if err != nil {
		return 0, empty
	}

	// 转化 Post
	postDTOs := make([]postdto.BriefDTO, 0, len(posts))
	for _, post := range posts {
		// 转成 DTO 返回给 Handler
		postDTO := postdto.ToBriefDTO(post, author)
		postDTOs = append(postDTOs, postDTO)
	}

	return int(total), postDTOs
}

// ListByPageAndTag 根据 Tag 分页查找帖子
func (svc *postService) ListByPageAndTag(ctx context.Context, name string, pageNo, pageSize int) (int, []postdto.DetailDTO) {
	var empty []postdto.DetailDTO

	tag, err := svc.tagRepo.GetByName(ctx, name)
	if err != nil {
		return 0, empty
	}
	// todo 避免性能问题，优化 SQL

	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPageAndTag(ctx, tag.ID, pageNo, pageSize)
	if err != nil {
		return 0, empty
	}

	// 转化
	var postDTOs []postdto.DetailDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		author, err := svc.userRepo.GetByID(ctx, post.UserID)
		if err != nil {
			slog.Warn("could not get name of user", "uid", post.UserID)
			author = &model.User{}
		}

		postDTO := postdto.ToDetailDTO(post, author)
		postDTOs = append(postDTOs, postDTO)
	}
	return int(total), postDTOs
}

// Like 点赞帖子
func (svc *postService) Like(ctx context.Context, pid, uid int64) error {
	// 查找帖子
	_, err := svc.postRepo.GetByID(ctx, pid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrPostNotFound
		}
		return errno.ErrServerInternal
	}

	// 创建点赞记录
	like := &model.Like{
		ID:     svc.idGen.NextID(),
		UserID: uid,
		PostID: pid,
	}
	err = svc.likeRepo.Like(ctx, like)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 重复点赞
			return errno.ErrDuplicatedLike
		}
		// 系统内部错误
		return errno.ErrServerInternal
	}

	field := model.PostLikeCount
	if err := svc.postRepo.UpdateCount(ctx, pid, field, 1); err != nil {
		slog.Error("Update Like Count Failed", "error", err)
	}

	return nil
}

// Unlike 取消点赞
func (svc *postService) Unlike(ctx context.Context, pid, uid int64) error {
	// 查找帖子
	_, err := svc.postRepo.GetByID(ctx, int64(pid))
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return errno.ErrPostNotFound
		}
		return errno.ErrServerInternal
	}

	// 删除点赞记录
	err = svc.likeRepo.UnLike(ctx, uid, pid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 重复删除
			return errno.ErrDuplicatedUnLike
		}
		// 系统内部错误
		return errno.ErrServerInternal
	}

	field := model.PostLikeCount
	if err := svc.postRepo.UpdateCount(ctx, pid, field, -1); err != nil {
		slog.Error("Update Like Count Failed", "error", err)
	}

	return nil
}

// IfLike 判断是否点过赞
func (svc *postService) IfLike(ctx context.Context, pid, uid int64) (bool, error) {
	ok, err := svc.likeRepo.HasLiked(ctx, uid, pid)
	if err != nil {
		return false, errno.ErrServerInternal
	}
	return ok, nil
}
