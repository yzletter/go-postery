package service

import (
	"context"
	"errors"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/response"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/utils"
)

var (
	ErrPostNotFound = errors.New("帖子不存在")
)

var Fields = []string{"view_count", "comment_count", "like_count"}

const (
	VIEW_CNT    = "view_count"
	COMMENT_CNT = "comment_count"
	LIKE_CNT    = "like_count"
)

type postService struct {
	PostRepo repository.PostRepository
	UserRepo repository.UserRepository
	LikeRepo repository.LikeRepository
	TagRepo  repository.TagRepository
	idGen    IDGenerator // 用于生成 ID
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository, likeRepo repository.LikeRepository, tagRepo repository.TagRepository) PostService {
	return &postService{
		PostRepo: postRepo,
		UserRepo: userRepo,
		LikeRepo: likeRepo,
		TagRepo:  tagRepo,
	}
}

func (svc *postService) Create(ctx context.Context, uid int, title, content string) (dto.PostDetailDTO, error) {
	post := &model.Post{
		ID:      svc.idGen.NextID(),
		UserID:  int64(uid),
		Title:   title,
		Content: content,
	}
	post, err := svc.PostRepo.Create(context.Background(), post)
	if err != nil {
		return dto.PostDetailDTO{}, err
	}
	user, _ := svc.UserRepo.GetByID(ctx, int64(uid))
	postDTO := dto.ToPostDetailDTO(*post, *user)
	return postDTO, err
}

func (svc *postService) Delete(ctx context.Context, pid, uid int) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(ctx, pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	// 删除帖子
	err := svc.PostRepo.Delete(ctx, int64(pid))
	return err
}

func (svc *postService) Update(ctx context.Context, pid int, uid int, title, content string, tags []string) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(ctx, pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	tagsBefore, err := svc.TagRepo.FindTagsByPostID(pid)
	if err != nil {
		slog.Error("Get Tags_Before Failed", "error", err)
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

	for _, tag := range tagsBefore {
		if _, ok := hashNow[tag]; !ok { // 原来有现在没有 ——> 删除
			tid, err := svc.TagRepo.Exist(tag) // 查 tid
			if err != nil {
				// 这里应该是必须有 tid 的才对
				slog.Error("Can Not Find Tid", "error", err)
			}
			err = svc.TagRepo.DeleteBind(pid, tid)
			if err != nil {
				slog.Error("Delete Bind Failed", "error", err)
			}
		}
	}

	for _, tag := range tagsNow {
		if _, ok := hashBefore[tag]; !ok { // 现在有原来没有 ——> 绑定
			tid, err := svc.TagRepo.Exist(tag) // 查 tid
			if err != nil {
				tid, err = svc.TagRepo.Create(tag, utils.Slugify(tag)) // 没有 tid 就新建
			}

			err = svc.TagRepo.Bind(pid, tid)
			if err != nil {

			}
		}
	}

	updates := map[string]any{
		"tittle":  title,
		"content": content,
	}
	err = svc.PostRepo.Update(ctx, int64(pid), updates) // 更新标题和正文
	return err
}

func (svc *postService) GetByPage(ctx context.Context, pageNo, pageSize int) (int, []dto.PostDetailDTO) {
	// 获取帖子总数和当前页帖子列表
	total, posts, _ := svc.PostRepo.GetByPage(ctx, pageNo, pageSize)

	var postDTOs []dto.PostDetailDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		user, err := svc.UserRepo.GetByID(ctx, post.UserID)
		if err != nil {
			slog.Warn("could not get name of user", "uid", post.UserID)
		} else {
			postDTO := dto.ToPostDetailDTO(*post, *user)
			postDTOs = append(postDTOs, postDTO)
		}
	}
	return int(total), postDTOs
}

func (svc *postService) GetByPageAndTag(ctx context.Context, name string, pageNo, pageSize int) (int, []dto.PostDetailDTO) {
	tid, err := svc.TagRepo.Exist(name)
	if err != nil {
		return 0, []dto.PostDetailDTO{}
	}

	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.PostRepo.GetByPageAndTag(ctx, int64(tid), pageNo, pageSize)

	var postDTOs []dto.PostDetailDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		user, err := svc.UserRepo.GetByID(ctx, post.UserID)
		if err != nil {
			slog.Warn("could not get name of user", "uid", post.UserID)
		} else {
			postDTO := dto.ToPostDetailDTO(*post, *user)
			postDTOs = append(postDTOs, postDTO)
		}
	}
	return int(total), postDTOs
}

func (svc *postService) GetDetailById(ctx context.Context, pid int) (bool, dto.PostDetailDTO) {
	post, err := svc.PostRepo.GetByID(ctx, int64(pid))
	if err != nil {
		return false, dto.PostDetailDTO{}
	}

	// 查找作者信息
	user, _ := svc.UserRepo.GetByID(ctx, post.UserID)

	// 记录 ViewCount + 1
	err = svc.PostRepo.UpdateCount(ctx, post.ID, 1, 1) // 数据库中 + 1

	post.ViewCount += 1
	postDTO := dto.ToPostDetailDTO(*post, *user)
	return true, postDTO
}

func (svc *postService) GetBriefById(ctx context.Context, pid int) (bool, dto.PostBriefDTO) {
	post, err := svc.PostRepo.GetByID(ctx, int64(pid))
	if err != nil {
		return false, dto.PostBriefDTO{}
	}

	// 查找作者信息
	user, _ := svc.UserRepo.GetByID(ctx, post.UserID)

	postBriefDTO := dto.ToPostBriefDTO(*post, *user)
	return true, postBriefDTO
}

func (svc *postService) HasMore(ctx context.Context, pageNo, pageSize, total int) bool {
	return pageNo*pageSize < total
}

// Belong 判断登录用户是否是帖子作者
func (svc *postService) Belong(ctx context.Context, pid, uid int) bool {
	ok, postDTO := svc.GetBriefById(ctx, pid)
	if !ok || uid != int(postDTO.Author.Id) {
		return false
	}
	return true
}

func (svc *postService) GetByUid(ctx context.Context, uid int) []dto.PostDetailDTO {
	_, posts, err := svc.PostRepo.GetByUid(ctx, int64(uid), 1, 10)
	if err != nil {
		return nil
	}

	postDTOs := make([]dto.PostDetailDTO, 0, len(posts))
	for _, post := range posts {
		// 查找作者信息
		user, _ := svc.UserRepo.GetByID(ctx, post.UserID)

		// 转成 DTO 返回给 Handler
		postDTO := dto.ToPostDetailDTO(*post, *user)
		postDTOs = append(postDTOs, postDTO)
	}

	return postDTOs
}

func (svc *postService) Like(ctx context.Context, pid, uid int) error {
	// 查找帖子
	_, err := svc.PostRepo.GetByID(ctx, int64(pid))
	if err != nil {
		// 帖子不存在
		return ErrPostNotFound
	}

	// 创建点赞记录
	err = svc.LikeRepo.Create(uid, pid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordHasExist) {
			// 重复点赞
			return repository.ErrRecordHasExist
		}
		// 系统内部错误
		return repository.ErrServerInternal
	}

	svc.PostRepo.UpdateCount(ctx, int64(pid), 3, 1)

	return nil
}

func (svc *postService) Dislike(ctx context.Context, pid, uid int) error {
	// 查找帖子
	_, err := svc.PostRepo.GetByID(ctx, int64(pid))
	if err != nil {
		// 帖子不存在
		return ErrPostNotFound
	}

	// 删除点赞记录
	err = svc.LikeRepo.Delete(uid, pid)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotExist) {
			// 重复删除
			return repository.ErrRecordNotExist
		}
		// 系统内部错误
		return repository.ErrServerInternal
	}

	svc.PostRepo.UpdateCount(ctx, int64(pid), 3, -1) // 数据库 + 1

	return nil
}

func (svc *postService) IfLike(pid, uid int) (bool, error) {
	ok, err := svc.LikeRepo.Get(uid, pid)
	if err != nil {
		return ok, repository.ErrServerInternal
	}
	return ok, nil
}
