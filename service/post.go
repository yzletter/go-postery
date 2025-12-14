package service

import (
	"errors"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/response"
	userLikeRepository "github.com/yzletter/go-postery/repository/like"
	postRepository "github.com/yzletter/go-postery/repository/post"
	tagRepository "github.com/yzletter/go-postery/repository/tag"
	userRepository "github.com/yzletter/go-postery/repository/user"
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

type PostService struct {
	PostDBRepo     *postRepository.PostDBRepository
	PostCacheRepo  *postRepository.PostCacheRepository
	UserDBRepo     *userRepository.UserDBRepository
	UserLikeDBRepo *userLikeRepository.UserLikeDBRepository
	TagDBRepo      *tagRepository.TagDBRepository
}

func NewPostService(postDBRepo *postRepository.PostDBRepository,
	postCacheRepo *postRepository.PostCacheRepository,
	userRepository *userRepository.UserDBRepository,
	userLikeDBRepo *userLikeRepository.UserLikeDBRepository,
	tagDBRepo *tagRepository.TagDBRepository,

) *PostService {
	return &PostService{
		PostDBRepo:     postDBRepo,
		PostCacheRepo:  postCacheRepo,
		UserDBRepo:     userRepository,
		UserLikeDBRepo: userLikeDBRepo,
		TagDBRepo:      tagDBRepo,
	}
}

func (svc *PostService) Create(uid int, title, content string) (dto.PostDetailDTO, error) {
	post, err := svc.PostDBRepo.Create(uid, title, content)
	_, user := svc.UserDBRepo.GetByID(uid)
	postDTO := dto.ToPostDetailDTO(post, user)
	return postDTO, err
}

func (svc *PostService) Delete(pid, uid int) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	// 删除帖子
	err := svc.PostDBRepo.Delete(pid)
	return err
}

func (svc *PostService) Update(pid int, uid int, title, content string, tags []string) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	tagsBefore, err := svc.TagDBRepo.FindTagsByPostID(pid)
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
			tid, err := svc.TagDBRepo.Exist(tag) // 查 tid
			if err != nil {
				// 这里应该是必须有 tid 的才对
				slog.Error("Can Not Find Tid", "error", err)
			}
			err = svc.TagDBRepo.DeleteBind(pid, tid)
			if err != nil {
				slog.Error("Delete Bind Failed", "error", err)
			}
		}
	}

	for _, tag := range tagsNow {
		if _, ok := hashBefore[tag]; !ok { // 现在有原来没有 ——> 绑定
			tid, err := svc.TagDBRepo.Exist(tag) // 查 tid
			if err != nil {
				tid, err = svc.TagDBRepo.Create(tag, utils.Slugify(tag)) // 没有 tid 就新建
			}

			err = svc.TagDBRepo.Bind(pid, tid)
			if err != nil {

			}
		}
	}

	err = svc.PostDBRepo.Update(pid, title, content) // 更新标题和正文
	return err
}

func (svc *PostService) GetByPage(pageNo, pageSize int) (int, []dto.PostDetailDTO) {
	// 获取帖子总数和当前页帖子列表
	total, posts := svc.PostDBRepo.GetByPage(pageNo, pageSize)
	var postDTOs []dto.PostDetailDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		ok, user := svc.UserDBRepo.GetByID(post.UserId)
		if !ok {
			slog.Warn("could not get name of user", "uid", post.UserId)
		}

		postDTO := dto.ToPostDetailDTO(post, user)
		postDTOs = append(postDTOs, postDTO)
	}
	return total, postDTOs
}

func (svc *PostService) GetByPageAndTag(name string, pageNo, pageSize int) (int, []dto.PostDetailDTO) {
	tid, err := svc.TagDBRepo.Exist(name)
	if err != nil {
		return 0, []dto.PostDetailDTO{}
	}

	// 获取帖子总数和当前页帖子列表
	total, posts := svc.PostDBRepo.GetByPageAndTag(tid, pageNo, pageSize)

	var postDTOs []dto.PostDetailDTO
	for _, post := range posts {
		// 根据 uid 找到 username 进行赋值
		ok, user := svc.UserDBRepo.GetByID(post.UserId)
		if !ok {
			slog.Warn("could not get name of user", "uid", post.UserId)
		}

		postDTO := dto.ToPostDetailDTO(post, user)
		postDTOs = append(postDTOs, postDTO)
	}
	return total, postDTOs
}

func (svc *PostService) GetDetailById(pid int) (bool, dto.PostDetailDTO) {
	ok, post := svc.PostDBRepo.GetByID(pid)
	if !ok {
		return false, dto.PostDetailDTO{}
	}

	// 查找作者信息
	_, user := svc.UserDBRepo.GetByID(post.UserId)

	// 记录 ViewCount + 1
	svc.PostDBRepo.ChangeViewCnt(post.Id, 1)                                // 数据库中 + 1
	ok, err := svc.PostCacheRepo.ChangeInteractiveCnt(VIEW_CNT, post.Id, 1) // 缓存中 + 1
	if !ok {                                                                // 缓存中没有 KEY
		vals := []int{post.ViewCount + 1, post.CommentCount, post.LikeCount}
		svc.PostCacheRepo.SetKey(pid, Fields, vals)
	}
	if err != nil {
		slog.Error("Redis Increase View Count Failed", "error", err)
	}

	post.ViewCount += 1
	postDTO := dto.ToPostDetailDTO(post, user)
	return true, postDTO
}

func (svc *PostService) GetBriefById(pid int) (bool, dto.PostBriefDTO) {
	ok, post := svc.PostDBRepo.GetByID(pid)
	if !ok {
		return false, dto.PostBriefDTO{}
	}

	// 查找作者信息
	_, user := svc.UserDBRepo.GetByID(post.UserId)

	postBriefDTO := dto.ToPostBriefDTO(post, user)
	return true, postBriefDTO
}

func (svc *PostService) HasMore(pageNo, pageSize, total int) bool {
	return pageNo*pageSize < total
}

// Belong 判断登录用户是否是帖子作者
func (svc *PostService) Belong(pid, uid int) bool {
	ok, postDTO := svc.GetBriefById(pid)
	if !ok || uid != int(postDTO.Author.Id) {
		return false
	}
	return true
}

func (svc *PostService) GetByUid(uid int) []dto.PostDetailDTO {
	posts := svc.PostDBRepo.GetByUid(uid)
	if posts == nil {
		return nil
	}

	postDTOs := make([]dto.PostDetailDTO, 0, len(posts))
	for _, post := range posts {
		// 查找作者信息
		_, user := svc.UserDBRepo.GetByID(post.UserId)

		// 转成 DTO 返回给 Handler
		postDTO := dto.ToPostDetailDTO(post, user)
		postDTOs = append(postDTOs, postDTO)
	}

	return postDTOs
}

func (svc *PostService) Like(pid, uid int) error {
	// 查找帖子
	ok, post := svc.PostDBRepo.GetByID(pid)
	if !ok {
		// 帖子不存在
		return ErrPostNotFound
	}

	// 创建点赞记录
	err := svc.UserLikeDBRepo.Create(uid, pid)
	if err != nil {
		if errors.Is(err, userLikeRepository.ErrRecordHasExist) {
			// 重复点赞
			return userLikeRepository.ErrRecordHasExist
		}
		// 系统内部错误
		return userRepository.ErrMySQLInternal
	}

	svc.PostDBRepo.ChangeLikeCnt(pid, 1)
	ok, err = svc.PostCacheRepo.ChangeInteractiveCnt(LIKE_CNT, pid, 1)
	if !ok {
		vals := []int{post.ViewCount, post.CommentCount, post.LikeCount + 1}
		svc.PostCacheRepo.SetKey(pid, Fields, vals)
	}

	return nil
}

func (svc *PostService) Dislike(pid, uid int) error {
	// 查找帖子
	ok, post := svc.PostDBRepo.GetByID(pid)
	if !ok {
		// 帖子不存在
		return ErrPostNotFound
	}

	// 删除点赞记录
	err := svc.UserLikeDBRepo.Delete(uid, pid)
	if err != nil {
		if errors.Is(err, userLikeRepository.ErrRecordNotExist) {
			// 重复删除
			return userLikeRepository.ErrRecordNotExist
		}
		// 系统内部错误
		return userRepository.ErrMySQLInternal
	}

	svc.PostDBRepo.ChangeLikeCnt(pid, -1)                               // 数据库 + 1
	ok, err = svc.PostCacheRepo.ChangeInteractiveCnt(LIKE_CNT, pid, -1) // 缓存 + 1
	if !ok {
		vals := []int{post.ViewCount, post.CommentCount, post.LikeCount - 1}
		svc.PostCacheRepo.SetKey(pid, Fields, vals)
	}

	return nil
}

func (svc *PostService) IfLike(pid, uid int) (bool, error) {
	ok, err := svc.UserLikeDBRepo.Get(uid, pid)
	if err != nil {
		return ok, userRepository.ErrMySQLInternal
	}
	return ok, nil
}
