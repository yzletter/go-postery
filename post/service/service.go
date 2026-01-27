package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bytedance/sonic"
	post_grpc "github.com/yzletter/go-postery/api/proto/post/v1"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/post/dto"
	"github.com/yzletter/go-postery/post/model"
	"github.com/yzletter/go-postery/post/repository"
	"github.com/yzletter/go-postery/post/service/ports"
	"github.com/yzletter/go-postery/post/utils"
)

type postService struct {
	postRepo repository.PostRepository
	likeRepo repository.LikeRepository
	tagRepo  repository.TagRepository
	idGen    ports.IDGenerator // 用于生成 ID
	post_grpc.UnimplementedPostServiceServer
}

func NewPostService(postRepo repository.PostRepository, likeRepo repository.LikeRepository, tagRepo repository.TagRepository, idGen ports.IDGenerator) PostService {
	return &postService{
		postRepo:                       postRepo,
		likeRepo:                       likeRepo,
		tagRepo:                        tagRepo,
		idGen:                          idGen,
		UnimplementedPostServiceServer: post_grpc.UnimplementedPostServiceServer{},
	}
}

func (svc *postService) Create(ctx context.Context, req *post_grpc.CreatePostRequest) (*post_grpc.PostDetail, error) {
	var empty = new(post_grpc.PostDetail)
	post := &model.Post{
		ID:          svc.idGen.NextID(),
		UserID:      req.UserID,
		Title:       req.Title,
		Content:     req.Content,
		ContentType: int(req.ContentType),
		Status:      1,
	}

	var payload model.ChunkDocumentEventPayload
	payload.ID = post.ID
	value, _ := sonic.MarshalString(payload)

	events := make([]*model.Event, 0)
	// 通知 RAG 新帖子建立
	event := &model.Event{
		ID:           svc.idGen.NextID(),
		Topic:        "index_document",
		MessageKey:   "index_document",
		MessageValue: value,
	}

	events = append(events, event)

	// 通知搜索引擎新帖子建立
	event = &model.Event{
		ID:           svc.idGen.NextID(),
		Topic:        "index_search",
		MessageKey:   "index_search",
		MessageValue: value,
	}
	events = append(events, event)

	if err := svc.postRepo.Create(ctx, post, events); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 雪花 ID 的帖子不会已存在, 需要排查
			slog.Error("Create Post Failed", "error", err)
		}
		return empty, errno.ErrServerInternal
	}

	_ = svc.BindTag(ctx, post.ID, req.Tags)
	return dto.ToPostDetail(post, req.Tags), nil
}

// GetDetailByID 获取帖子详情，并选择是否增加浏览量
func (svc *postService) GetDetailByID(ctx context.Context, req *post_grpc.GetDetailByIDRequest) (*post_grpc.PostDetail, error) {
	var empty = new(post_grpc.PostDetail)

	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrPostNotFound
		}
		return empty, errno.ErrServerInternal
	}

	if req.AddViewCnt {
		// 记录 ViewCount + 1
		if err = svc.postRepo.UpdateCount(ctx, post.ID, model.PostViewCount, 1); err != nil {
			if errors.Is(err, repository.ErrRecordNotFound) {
				slog.Error("Update View Cnt Failed", "error", err)
			}
		}
		post.ViewCount += 1
	}

	// 查询 Tags
	tags, _ := svc.FindTagsByPostID(ctx, post.ID)

	return dto.ToPostDetail(post, tags), nil
}

// GetBriefByID 根据 ID 获取帖子简要信息
func (svc *postService) GetBriefByID(ctx context.Context, req *post_grpc.GetBriefByIDRequest) (*post_grpc.PostBrief, error) {
	var empty = new(post_grpc.PostBrief)

	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return empty, errno.ErrPostNotFound
		}
		return empty, errno.ErrServerInternal
	}

	return dto.ToPostBrief(post), nil
}

func (svc *postService) Top(ctx context.Context, req *post_grpc.PostEmptyRequest) (*post_grpc.TopResponse, error) {
	var empty = new(post_grpc.TopResponse)

	posts, scores, err := svc.postRepo.Top(ctx)
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var topPosts []*post_grpc.TopPost
	for k, post := range posts {
		topPost := dto.ToTopPost(post, scores[k])
		topPosts = append(topPosts, topPost)
	}

	return &post_grpc.TopResponse{TopPosts: topPosts}, nil
}

// Update 更新帖子
func (svc *postService) Update(ctx context.Context, req *post_grpc.UpdateRequest) (*post_grpc.PostEmptyResponse, error) { // 判断登录用户是否是作者
	// 判断是否是作者
	post, err := svc.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &post_grpc.PostEmptyResponse{}, errno.ErrPostNotFound
		}
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	} else if post.UserID != req.UserID {
		return &post_grpc.PostEmptyResponse{}, errno.ErrUnauthorized
	}

	tagsBefore, err := svc.tagRepo.FindTagsByPostID(ctx, req.PostID)
	if err != nil {
		slog.Error("Get Tags_Before Failed", "error", err)
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}

	tagsNow := req.Tags

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
			if err = svc.tagRepo.DeleteBind(ctx, req.PostID, tag.ID); err != nil {
				slog.Error("DeleteScore BindTag Failed", "error", err)
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
				PostID: req.PostID,
				TagID:  tag.ID,
			}
			if err = svc.tagRepo.Bind(ctx, postTag); err != nil {
				slog.Error("BindTag Post Tag Failed", "error", err)
			}
		}
	}

	updates := map[string]any{
		"title":   req.Title,
		"content": req.Content,
	}

	// 更新标题和正文
	if err = svc.postRepo.Update(ctx, req.PostID, updates); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &post_grpc.PostEmptyResponse{}, errno.ErrPostNotFound
		}
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}

	return &post_grpc.PostEmptyResponse{}, nil
}

// ListByPage 按页获取帖子列表
func (svc *postService) ListByPage(ctx context.Context, req *post_grpc.ListByPageRequest) (*post_grpc.PostDetailsResponse, error) {
	var empty = new(post_grpc.PostDetailsResponse)

	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPage(ctx, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return empty, errno.ErrPostNotFound
	}

	// todo 避免性能问题，优化 SQL
	// 转化
	var postDetails []*post_grpc.PostDetail
	for _, post := range posts {
		// 查找当前帖子的 Tags
		tags, _ := svc.FindTagsByPostID(ctx, post.ID)
		postDetail := dto.ToPostDetail(post, tags)
		postDetails = append(postDetails, postDetail)
	}
	return &post_grpc.PostDetailsResponse{
		Count:       uint64(total),
		PostDetails: postDetails,
	}, nil
}

// ListByPageAndUid 根据作者 ID 获取帖子简要信息列表
func (svc *postService) ListByPageAndUid(ctx context.Context, req *post_grpc.ListByPageAndUidRequest) (*post_grpc.PostBriefsResponse, error) {
	var empty = new(post_grpc.PostBriefsResponse)
	total, posts, err := svc.postRepo.GetByUid(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return empty, errno.ErrPostNotFound
	}

	// 转化 Post
	var postBriefs []*post_grpc.PostBrief
	for _, post := range posts {
		// 转成 DTO 返回给 Handler
		postBrief := dto.ToPostBrief(post)
		postBriefs = append(postBriefs, postBrief)
	}

	return &post_grpc.PostBriefsResponse{
		Count:      uint64(total),
		PostBriefs: postBriefs,
	}, nil

}

// ListByPageAndTag 根据 Tag 分页查找帖子
func (svc *postService) ListByPageAndTag(ctx context.Context, req *post_grpc.ListByPageAndTagRequest) (*post_grpc.PostDetailsResponse, error) {
	var empty = new(post_grpc.PostDetailsResponse)

	tag, err := svc.tagRepo.GetByName(ctx, req.Tag)
	if err != nil {
		return empty, errno.ErrPostNotFound
	}

	// todo 避免性能问题，优化 SQL
	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPageAndTag(ctx, tag.ID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return empty, errno.ErrPostNotFound
	}

	// 转化
	var postDetails []*post_grpc.PostDetail
	for _, post := range posts {
		// 查找当前帖子的 Tags
		tags, _ := svc.FindTagsByPostID(ctx, post.ID)
		postDetail := dto.ToPostDetail(post, tags)
		postDetails = append(postDetails, postDetail)
	}
	return &post_grpc.PostDetailsResponse{
		Count:       uint64(total),
		PostDetails: postDetails,
	}, nil
}

// Belong 判断登录用户是否是帖子作者
func (svc *postService) Belong(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.BelongResponse, error) {
	// todo 优化只查 user_id 字段
	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &post_grpc.BelongResponse{Result: false}, errno.ErrPostNotFound
		}
		return &post_grpc.BelongResponse{Result: false}, errno.ErrServerInternal
	} else if post.UserID != req.UserID {
		return &post_grpc.BelongResponse{Result: false}, nil
	}
	return &post_grpc.BelongResponse{Result: true}, nil
}

// Delete 删除帖子
func (svc *postService) Delete(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	// 判断登录用户是否是作者
	post, err := svc.postRepo.GetByID(ctx, req.PostID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &post_grpc.PostEmptyResponse{}, errno.ErrPostNotFound
		}
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	} else if post.UserID != req.UserID {
		return &post_grpc.PostEmptyResponse{}, errno.ErrUnauthorized
	}

	// 删除帖子
	if err := svc.postRepo.Delete(ctx, req.PostID); err != nil {
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}
	return &post_grpc.PostEmptyResponse{}, nil
}

func (svc *postService) Like(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	// 查找帖子
	if _, err := svc.postRepo.GetByID(ctx, req.PostID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &post_grpc.PostEmptyResponse{}, errno.ErrPostNotFound
		}
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}

	// 创建点赞记录
	like := &model.Like{
		ID:     svc.idGen.NextID(),
		UserID: req.UserID,
		PostID: req.PostID,
	}

	if err := svc.likeRepo.Like(ctx, like); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			// 重复点赞
			return &post_grpc.PostEmptyResponse{}, errno.ErrDuplicatedLike
		}
		// 系统内部错误
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}

	// 修改分数
	svc.postRepo.ChangeScore(ctx, req.PostID, 432)

	if err := svc.postRepo.UpdateCount(ctx, req.PostID, model.PostLikeCount, 1); err != nil {
		slog.Error("Update Like Count Failed", "error", err)
	}

	return &post_grpc.PostEmptyResponse{}, nil
}

func (svc *postService) Unlike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.PostEmptyResponse, error) {
	// 查找帖子
	if _, err := svc.postRepo.GetByID(ctx, req.PostID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return &post_grpc.PostEmptyResponse{}, errno.ErrPostNotFound
		}
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}

	// 删除点赞记录
	if err := svc.likeRepo.UnLike(ctx, req.UserID, req.PostID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			// 重复删除
			return &post_grpc.PostEmptyResponse{}, errno.ErrDuplicatedUnLike
		}
		// 系统内部错误
		return &post_grpc.PostEmptyResponse{}, errno.ErrServerInternal
	}

	// 修改分数
	svc.postRepo.ChangeScore(ctx, req.PostID, -432)
	if err := svc.postRepo.UpdateCount(ctx, req.PostID, model.PostLikeCount, -1); err != nil {
		slog.Error("Update Like Count Failed", "error", err)
	}

	return &post_grpc.PostEmptyResponse{}, nil
}

func (svc *postService) IfLike(ctx context.Context, req *post_grpc.PostCommonRequest) (*post_grpc.IfLikeResponse, error) {
	if ok, err := svc.likeRepo.HasLiked(ctx, req.UserID, req.PostID); err == nil {
		return &post_grpc.IfLikeResponse{Result: ok}, nil
	}
	return &post_grpc.IfLikeResponse{Result: false}, errno.ErrServerInternal
}

// BindTag 将 Tags 绑定到 post
func (svc *postService) BindTag(ctx context.Context, pid int64, tags []string) error {
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
			slog.Error("BindTag Post Tag Failed", "error", err)
			if errors.Is(err, repository.ErrUniqueKey) {
				return errno.ErrTagDuplicatedBind
			}
			return errno.ErrServerInternal
		}
	}

	return nil
}

// CreateTag 新建 Tag
func (svc *postService) CreateTag(ctx context.Context, name string) (int64, error) {
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

// FindTagsByPostID 根据帖子 ID 查找 Tag
func (svc *postService) FindTagsByPostID(ctx context.Context, pid int64) ([]string, error) {
	var empty []string
	res, err := svc.tagRepo.FindTagsByPostID(ctx, pid)
	if err != nil {
		slog.Error("Find Tags By Post ID Failed", "error", err)
		return empty, errno.ErrServerInternal
	}
	return res, nil
}
