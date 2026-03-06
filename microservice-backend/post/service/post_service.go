package service

import (
	"context"
	"errors"
	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/yzletter/go-postery/microservice-backend/post/errs"
	model2 "github.com/yzletter/go-postery/microservice-backend/post/model"
	repository2 "github.com/yzletter/go-postery/microservice-backend/post/repository"
	"github.com/yzletter/go-postery/microservice-backend/post/service/ports"
	"github.com/yzletter/go-postery/microservice-backend/post/utils"
)

type postService struct {
	postRepo    repository2.PostRepository
	likeRepo    repository2.LikeRepository
	tagRepo     repository2.TagRepository
	commentRepo repository2.CommentRepository
	idGen       ports.IDGenerator // 用于生成 ID
}

func NewPostService(postRepo repository2.PostRepository, likeRepo repository2.LikeRepository, tagRepo repository2.TagRepository, commentRepo repository2.CommentRepository, idGen ports.IDGenerator) PostService {
	return &postService{
		postRepo:    postRepo,
		likeRepo:    likeRepo,
		tagRepo:     tagRepo,
		commentRepo: commentRepo,
		idGen:       idGen,
	}
}

func (svc *postService) Create(ctx context.Context, userID int64, title string, content string, contentType int, tags []string) (*PostWithTags, error) {
	post := &model2.Post{
		ID:          svc.idGen.NextID(),
		UserID:      userID,
		Title:       title,
		Content:     content,
		ContentType: contentType,
		Status:      1,
	}

	var payload model2.ChunkDocumentEventPayload
	payload.ID = post.ID
	value, _ := sonic.MarshalString(payload)

	events := make([]*model2.Event, 0)
	// 通知 RAG 新帖子建立
	event := &model2.Event{
		ID:           svc.idGen.NextID(),
		Topic:        "index_document",
		MessageKey:   "index_document",
		MessageValue: value,
	}

	events = append(events, event)

	// 通知搜索引擎新帖子建立
	event = &model2.Event{
		ID:           svc.idGen.NextID(),
		Topic:        "index_search",
		MessageKey:   "index_search",
		MessageValue: value,
	}
	events = append(events, event)

	if err := svc.postRepo.Create(ctx, post, events); err != nil {
		if errors.Is(err, repository2.ErrUniqueKey) {
			// 雪花 ID 的帖子不会已存在, 需要排查
			slog.Error("Create Post Failed", "error", err)
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	_ = svc.BindTag(ctx, post.ID, tags)
	return &PostWithTags{Post: post, Tags: tags}, nil
}

// GetDetailByID 获取帖子详情，并选择是否增加浏览量
func (svc *postService) GetDetailByID(ctx context.Context, postID int64, addViewCnt bool) (*PostWithTags, error) {
	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found", "error", err)
			return nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	if addViewCnt {
		// 记录 ViewCount + 1
		if err = svc.postRepo.UpdateCount(ctx, post.ID, model2.PostViewCount, 1); err != nil {
			if errors.Is(err, repository2.ErrRecordNotFound) {
				slog.Error("Update View Cnt Failed", "error", err)
			}
		}
		post.ViewCount += 1
	}

	// 查询 Tags
	tags, _ := svc.FindTagsByPostID(ctx, post.ID)

	return &PostWithTags{Post: post, Tags: tags}, nil
}

// GetBriefByID 根据 ID 获取帖子简要信息
func (svc *postService) GetBriefByID(ctx context.Context, postID int64) (*model2.Post, error) {
	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found", "error", err)
			return nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	return post, nil
}

func (svc *postService) Top(ctx context.Context) ([]*model2.Post, []float64, error) {
	posts, scores, err := svc.postRepo.Top(ctx)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return nil, nil, errs.ErrInternal
	}

	return posts, scores, nil
}

// Update 更新帖子
func (svc *postService) Update(ctx context.Context, userID int64, postID int64, title string, content string, tags []string) error {
	// 判断是否是作者
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	} else if post.UserID != userID { // 没有权利修改
		slog.Error("Unauthenticated")
		return errs.ErrUnauthenticated
	}

	tagsBefore, err := svc.tagRepo.FindTagsByPostID(ctx, postID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
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
			if err = svc.tagRepo.DeleteBind(ctx, postID, tag.ID); err != nil {
				slog.Error("DeleteScore BindTag Failed", "error", err)
			}
		}
	}

	for _, tagName := range tagsNow {
		if _, ok := hashBefore[tagName]; !ok { // 现在有原来没有 ——> 绑定
			// todo GetOrCreateByName 并发
			tag, err := svc.tagRepo.GetByName(ctx, tagName) // 查 tid
			if err != nil {
				newTag := &model2.Tag{
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
			postTag := &model2.PostTag{
				ID:     svc.idGen.NextID(),
				PostID: postID,
				TagID:  tag.ID,
			}
			if err = svc.tagRepo.Bind(ctx, postTag); err != nil {
				slog.Error("BindTag Post Tag Failed", "error", err)
			}
		}
	}

	updates := map[string]any{
		"title":   title,
		"content": content,
	}

	// 更新标题和正文
	if err = svc.postRepo.Update(ctx, postID, updates); err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	return nil
}

// ListByPage 按页获取帖子列表
func (svc *postService) ListByPage(ctx context.Context, pageNo int, pageSize int) (int64, []*PostWithTags, error) {
	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPage(ctx, pageNo, pageSize)
	if err != nil {
		slog.Error("Post Not Found")
		return 0, nil, errs.ErrNotFound
	}

	// todo 避免性能问题，优化 SQL
	postDetails := make([]*PostWithTags, 0, len(posts))
	for _, post := range posts {
		// 查找当前帖子的 Tags
		tags, _ := svc.FindTagsByPostID(ctx, post.ID)
		postDetails = append(postDetails, &PostWithTags{Post: post, Tags: tags})
	}
	return total, postDetails, nil
}

// ListByPageAndUid 根据作者 ID 获取帖子简要信息列表
func (svc *postService) ListByPageAndUid(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []*model2.Post, error) {
	total, posts, err := svc.postRepo.GetByUid(ctx, userID, pageNo, pageSize)
	if err != nil {
		slog.Error("Post Not Found")
		return 0, nil, errs.ErrNotFound
	}

	return total, posts, nil

}

// ListByPageAndTag 根据 Tag 分页查找帖子
func (svc *postService) ListByPageAndTag(ctx context.Context, tagName string, pageNo int, pageSize int) (int64, []*PostWithTags, error) {
	tag, err := svc.tagRepo.GetByName(ctx, tagName)
	if err != nil {
		slog.Error("Post Not Found")
		return 0, nil, errs.ErrNotFound
	}

	// todo 避免性能问题，优化 SQL
	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPageAndTag(ctx, tag.ID, pageNo, pageSize)
	if err != nil {
		slog.Error("Post Not Found")
		return 0, nil, errs.ErrNotFound
	}

	postDetails := make([]*PostWithTags, 0, len(posts))
	for _, post := range posts {
		// 查找当前帖子的 Tags
		tags, _ := svc.FindTagsByPostID(ctx, post.ID)
		postDetails = append(postDetails, &PostWithTags{Post: post, Tags: tags})
	}
	return total, postDetails, nil
}

// Belong 判断登录用户是否是帖子作者
func (svc *postService) Belong(ctx context.Context, userID int64, postID int64) (bool, error) {
	// todo 优化只查 user_id 字段
	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return false, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return false, errs.ErrInternal
	} else if post.UserID != userID {
		return false, nil
	}
	return true, nil
}

// Delete 删除帖子
func (svc *postService) Delete(ctx context.Context, userID int64, postID int64) error {
	// 判断登录用户是否是作者
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	} else if post.UserID != userID {
		slog.Error("Unauthenticated")
		return errs.ErrUnauthenticated
	}

	// 删除帖子
	if err := svc.postRepo.Delete(ctx, postID); err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}
	return nil
}

func (svc *postService) Like(ctx context.Context, userID int64, postID int64) error {
	// 查找帖子
	if _, err := svc.postRepo.GetByID(ctx, postID); err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 创建点赞记录
	like := &model2.Like{
		ID:     svc.idGen.NextID(),
		UserID: userID,
		PostID: postID,
	}

	if err := svc.likeRepo.Like(ctx, like); err != nil {
		if errors.Is(err, repository2.ErrUniqueKey) {
			// 重复点赞
			slog.Error("Duplicated Like")
			return errs.ErrAlreadyExits
		}
		// 系统内部错误
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 修改分数
	svc.postRepo.ChangeScore(ctx, postID, 432)

	if err := svc.postRepo.UpdateCount(ctx, postID, model2.PostLikeCount, 1); err != nil {
		slog.Error("Update Like Count Failed", "error", err)
	}

	return nil
}

func (svc *postService) Unlike(ctx context.Context, userID int64, postID int64) error {
	// 查找帖子
	if _, err := svc.postRepo.GetByID(ctx, postID); err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 删除点赞记录
	if err := svc.likeRepo.UnLike(ctx, userID, postID); err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			// 重复删除
			slog.Error("Duplicated Unlike")
			return errs.ErrAlreadyExits
		}
		// 系统内部错误
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 修改分数
	svc.postRepo.ChangeScore(ctx, postID, -432)
	if err := svc.postRepo.UpdateCount(ctx, postID, model2.PostLikeCount, -1); err != nil {
		slog.Error("Update Like Count Failed", "error", err)
	}

	return nil
}

func (svc *postService) IfLike(ctx context.Context, userID int64, postID int64) (bool, error) {
	if ok, err := svc.likeRepo.HasLiked(ctx, userID, postID); err == nil {
		return ok, nil
	} else {
		slog.Error("Server Internal Error", "error", err)
		return false, errs.ErrInternal
	}
}

func (svc *postService) CreateComment(ctx context.Context, postID int64, parentID int64, replyID int64, userID int64, content string) (*model2.Comment, error) {
	// 查询帖子
	_, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	// 新建评论
	comment := &model2.Comment{
		ID:       svc.idGen.NextID(),
		PostID:   postID,
		ParentID: parentID,
		ReplyID:  replyID,
		UserID:   userID,
		Content:  content,
	}

	if err := svc.commentRepo.Create(ctx, comment); err != nil {
		if errors.Is(err, repository2.ErrUniqueKey) {
			// 雪花 ID 的评论不会已存在, 需要排查
			slog.Error("Create Comment Failed", "error", err)
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	// 修改评论数
	if err := svc.postRepo.UpdateCount(ctx, postID, model2.PostCommentCount, 1); err != nil {
		slog.Error("Update Comment Count Failed", "error", err)
	}

	return comment, err
}

func (svc *postService) DeleteComment(ctx context.Context, commentID int64, userID int64) error {
	// 判断是否有删除权限
	comment, err := svc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Comment Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	if comment.UserID != userID {
		slog.Error("ErrUnauthenticated")
		return errs.ErrUnauthenticated
	}

	// 删除评论
	cnt, err := svc.commentRepo.Delete(ctx, commentID) // 返回被删除的个数
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Comment Not Found")
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	// 改变评论数

	if err := svc.postRepo.UpdateCount(ctx, comment.PostID, model2.PostCommentCount, -cnt); err != nil {
		slog.Error("Update Comment Failed", "error", err)
	}

	return nil
}

// ListCommentByPage 根据 PostID 按页获取文章主评论
func (svc *postService) ListCommentByPage(ctx context.Context, postID int64, pageNo int, pageSize int) (int64, []*model2.Comment, error) {
	total, comments, err := svc.commentRepo.GetByPostID(ctx, postID, pageNo, pageSize)
	if err != nil {
		slog.Error("Comment Not Found")
		return 0, nil, errs.ErrNotFound
	}

	return total, comments, nil
}

// ListRepliesByPage 根据 CommentID 按页获取评论的回复
func (svc *postService) ListRepliesByPage(ctx context.Context, commentID int64, pageNo int, pageSize int) (int64, []*model2.Comment, error) {
	total, comments, err := svc.commentRepo.GetRepliesByParentID(ctx, commentID, pageNo, pageSize)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Comment Not Found")
			return 0, nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return 0, nil, errs.ErrInternal
	}

	return total, comments, nil
}

// CheckCommentDeleteAuth 评论是否属于用户
func (svc *postService) CheckCommentDeleteAuth(ctx context.Context, commentID int64, userID int64) (bool, error) {
	comment, err := svc.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Comment Not Found")
			return false, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return false, errs.ErrInternal
	}

	post, err := svc.postRepo.GetByID(ctx, comment.PostID)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Post Not Found")
			return false, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return false, errs.ErrInternal
	}

	// 帖子属于当前登录用户，或评论属于当前用户
	return comment.UserID == userID || post.UserID == userID, nil
}

// BindTag 将 Tags 绑定到 post
func (svc *postService) BindTag(ctx context.Context, pid int64, tags []string) error {
	for _, tagName := range tags {
		// todo GetOrCreateByName 并发
		tag, err := svc.tagRepo.GetByName(ctx, tagName) // 查 tid
		if err != nil {
			newTag := &model2.Tag{
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
		postTag := &model2.PostTag{
			ID:     svc.idGen.NextID(),
			PostID: pid,
			TagID:  tag.ID,
		}
		if err = svc.tagRepo.Bind(ctx, postTag); err != nil {
			slog.Error("BindTag Post Tag Failed", "error", err)
			if errors.Is(err, repository2.ErrUniqueKey) {
				slog.Error("Duplicated Tag Bind")
				return errs.ErrAlreadyExits
			}
			slog.Error("Server Internal Error", "error", err)
			return errs.ErrInternal
		}
	}

	return nil
}

// CreateTag 新建 Tag
func (svc *postService) CreateTag(ctx context.Context, name string) (int64, error) {
	// 获得唯一标识符
	tagName := name
	slug := utils.Slugify(name)

	tag := &model2.Tag{
		ID:   svc.idGen.NextID(),
		Name: tagName,
		Slug: slug,
	}

	if err := svc.tagRepo.Create(ctx, tag); err != nil {
		if !errors.Is(err, repository2.ErrUniqueKey) {
			slog.Error("Server Internal Error", "error", err)
			return 0, errs.ErrInternal
		}
	}
	return tag.ID, nil
}

// FindTagsByPostID 根据帖子 ID 查找 Tag
func (svc *postService) FindTagsByPostID(ctx context.Context, pid int64) ([]string, error) {
	var empty []string
	res, err := svc.tagRepo.FindTagsByPostID(ctx, pid)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return empty, errs.ErrInternal
	}
	return res, nil
}
