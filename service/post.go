package service

import (
	"errors"
	"log/slog"

	dto "github.com/yzletter/go-postery/dto/response"
	userLikeRepository "github.com/yzletter/go-postery/repository/like"
	postRepository "github.com/yzletter/go-postery/repository/post"
	userRepository "github.com/yzletter/go-postery/repository/user"
)

var (
	ErrPostNotFound = errors.New("帖子不存在")
)

type PostService struct {
	PostDBRepo     *postRepository.PostDBRepository
	PostCacheRepo  *postRepository.PostCacheRepository
	UserDBRepo     *userRepository.UserDBRepository
	UserLikeDBRepo *userLikeRepository.UserLikeDBRepository
}

func NewPostService(postDBRepo *postRepository.PostDBRepository,
	postCacheRepo *postRepository.PostCacheRepository,
	userRepository *userRepository.UserDBRepository,
	userLikeDBRepo *userLikeRepository.UserLikeDBRepository,
) *PostService {
	return &PostService{
		PostDBRepo:     postDBRepo,
		PostCacheRepo:  postCacheRepo,
		UserDBRepo:     userRepository,
		UserLikeDBRepo: userLikeDBRepo,
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

func (svc *PostService) Update(pid int, uid int, title, content string) error {
	// 判断登录用户是否是作者
	ok := svc.Belong(pid, uid)
	if !ok {
		// 无权限删除
		return errors.New("没有权限")
	}

	err := svc.PostDBRepo.Update(pid, title, content)
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

func (svc *PostService) GetDetailById(pid int) (bool, dto.PostDetailDTO) {
	ok, post := svc.PostDBRepo.GetByID(pid)
	if !ok {
		return false, dto.PostDetailDTO{}
	}

	// 查找作者信息
	_, user := svc.UserDBRepo.GetByID(post.UserId)

	// 记录 ViewCount + 1
	svc.PostDBRepo.ChangeViewCnt(post.Id, 1)               // 数据库中 + 1
	ok, err := svc.PostCacheRepo.ChangeViewCnt(post.Id, 1) // 缓存中 + 1
	if !ok {                                               // 缓存中没有 KEY
		svc.PostCacheRepo.SetKey(post.Id, "view_cnt", post.ViewCount+1)
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
	if !ok || uid != postDTO.Author.Id {
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
	ok, err = svc.PostCacheRepo.ChangeLikeCnt(pid, 1)
	if !ok {
		svc.PostCacheRepo.SetKey(pid, "like_cnt", post.LikeCount+1)
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

	svc.PostDBRepo.ChangeLikeCnt(pid, -1)              // 数据库 + 1
	ok, err = svc.PostCacheRepo.ChangeLikeCnt(pid, -1) // 缓存 + 1
	if !ok {
		svc.PostCacheRepo.SetKey(pid, "like_cnt", post.LikeCount-1)
	}

	return nil
}
