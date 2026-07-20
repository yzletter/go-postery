package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	interactive_grpc "github.com/yzletter/go-postery/api/proto/interactive/v1"
	rank_grpc "github.com/yzletter/go-postery/api/proto/rank/v1"
	"github.com/yzletter/go-postery/backend/event"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/post/domain"
	"github.com/yzletter/go-postery/backend/micro/post/model"
	"github.com/yzletter/go-postery/backend/micro/post/repository"
	"github.com/yzletter/go-postery/backend/ports"
	"github.com/yzletter/go-postery/backend/utils"
	"golang.org/x/sync/errgroup"
)

type postService struct {
	postRepo          repository.PostRepository
	tagRepo           repository.TagRepository
	interactiveClient manager.InteractiveClient
	rankClient        manager.RankClient
	idGen             ports.IDGenerator // 用于生成 ID
}

func NewPostService(postRepo repository.PostRepository, tagRepo repository.TagRepository, idGen ports.IDGenerator,
	interServiceClient manager.InteractiveClient, rankServiceClient manager.RankClient) PostService {
	return &postService{
		postRepo:          postRepo,
		tagRepo:           tagRepo,
		idGen:             idGen,
		interactiveClient: interServiceClient,
		rankClient:        rankServiceClient,
	}
}

// Create 创建新帖子并通知其他微服务
func (svc *postService) Create(ctx context.Context, post domain.Post) (domain.Post, error) {
	post.ID = svc.idGen.NextID()
	if post.Status == 0 {
		post.Status = 1
	}

	// 构造 Outbox
	events := make([]*event.OutboxEvent, 0)

	payload := event.NewPostEventPayload{ID: post.ID}

	// Search 事件
	searchEvent := event.NewKafkaOutboxEvent(svc.idGen.NextID(), event.KafkaSearchTopic, event.KafkaSearchGroup, payload)
	events = append(events, searchEvent)

	// 创建帖子、绑定标签并写 Outbox
	if err := svc.postRepo.Create(ctx, post, events); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) {
			slog.Warn("create post id conflict", "id", post.ID, "error", err)
		} else {
			slog.Error("create post failed", "id", post.ID, "error", err)
		}
		return domain.Post{}, errs.ErrInternal
	}

	// 查询 MySQL 自动生成的时间
	if createdPost, err := svc.postRepo.GetByID(ctx, post.ID); err == nil {
		post.CreatedAt = createdPost.CreatedAt
		post.UpdatedAt = createdPost.UpdatedAt
	} else {
		slog.Warn("get created post failed", "id", post.ID, "error", err)
	}

	// 异步初始化新帖子榜单分数
	if svc.rankClient != nil {
		go func() {
			rankCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_, _ = svc.rankClient.RankPost(rankCtx, &rank_grpc.RankIDRequest{ID: post.ID})
		}()
	}

	// 查询 Tag
	post.Tags, _ = svc.FindTagsByPostID(ctx, post.ID)

	return post, nil
}

// GetDetailByID 获取帖子详情，并选择是否增加浏览量
func (svc *postService) GetDetailByID(ctx context.Context, postID int64, addViewCnt bool) (domain.Post, error) {
	var post domain.Post
	var tags []string
	var err error
	var viewCnt int64
	var likeCnt int64
	var commentCnt int64
	var interactiveOK bool

	var eg errgroup.Group
	// 异步查找帖子详情
	eg.Go(func() error {
		var eerr error
		post, eerr = svc.postRepo.GetByID(ctx, postID)
		if eerr != nil {
			if errors.Is(eerr, repository.ErrRecordNotFound) {
				slog.Info("post not found", "id", postID)
				return errs.ErrNotFound
			}
			slog.Error("get post by id failed", "id", postID, "error", eerr)
			return errs.ErrInternal
		}
		return nil
	})

	// 异步查询 Tag
	eg.Go(func() error {
		var eerr error
		tags, eerr = svc.FindTagsByPostID(ctx, postID)
		if eerr != nil {
			slog.Warn("find post tags failed", "id", postID, "error", eerr)
		}
		return nil
	})

	// 异步查询 Interactive
	eg.Go(func() error {
		if svc.interactiveClient == nil {
			return nil
		}
		resp, eerr := svc.interactiveClient.GetPostInteractive(ctx, &interactive_grpc.PostIDRequest{PostID: postID})
		if eerr != nil {
			slog.Warn("get post interactive failed", "post_id", postID, "error", eerr)
			return nil
		}
		viewCnt = resp.ReadCnt
		likeCnt = resp.LikeCnt
		commentCnt = resp.CommentCnt
		interactiveOK = true
		return nil
	})

	if err = eg.Wait(); err != nil {
		return domain.Post{}, err
	}

	if addViewCnt {
		// todo 发消息加阅读
		//// 记录 ViewCount + 1
		//if err = svc.postRepo.UpdateCount(ctx, post.ID, model.PostViewCount, 1); err != nil {
		//	if errors.Is(err, repository.ErrRecordNotFound) {
		//		slog.Error("Update View Cnt Failed", "error", err)
		//	}
		//}
		//post.ViewCount += 1
	}

	post.Tags = tags
	if interactiveOK {
		post.ViewCount = int(viewCnt)
		post.LikeCount = int(likeCnt)
		post.CommentCount = int(commentCnt)
	}

	return post, nil
}

// GetBriefByID 根据帖子 ID 获取帖子简要信息
func (svc *postService) GetBriefByID(ctx context.Context, postID int64) (domain.PostBrief, error) {
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("post not found", "id", postID)
			return domain.PostBrief{}, errs.ErrNotFound
		}
		slog.Error("get post brief by id failed", "id", postID, "error", err)
		return domain.PostBrief{}, errs.ErrInternal
	}
	return post.Briefed(), nil
}

// GetPostByTime 根据时间获取帖子 ID
func (svc *postService) GetPostByTime(ctx context.Context, timeAt time.Time) ([]int64, error) {
	ids, err := svc.postRepo.GetPostByTime(ctx, timeAt)
	if err != nil {
		slog.Error("get post by time failed", "error", err)
		return nil, errs.ErrInternal
	}
	return ids, nil
}

func (svc *postService) Top(ctx context.Context) ([]domain.PostBrief, []float64, error) {
	if svc.rankClient == nil {
		return nil, nil, errs.ErrUnavailable
	}

	resp, err := svc.rankClient.TopKPost(ctx, &rank_grpc.RankEmptyRequest{})
	if err != nil {
		slog.Error("get top posts failed", "error", err)
		return nil, nil, errs.ErrInternal
	}

	posts := make([]domain.PostBrief, 0, len(resp.Posts))
	scores := make([]float64, 0, len(resp.Posts))
	for _, rankedPost := range resp.Posts {
		postDetail, err := svc.GetBriefByID(ctx, rankedPost.ID)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				posts = append(posts, domain.PostBrief{
					ID:    rankedPost.ID,
					Title: "未知文章",
				})
				scores = append(scores, float64(rankedPost.Score))
				continue
			}
			return nil, nil, err
		}
		posts = append(posts, postDetail)
		scores = append(scores, float64(rankedPost.Score))
	}

	return posts, scores, nil
}

// Update 更新帖子
func (svc *postService) Update(ctx context.Context, updatePost domain.Post) error {
	userID := updatePost.User.UserID
	postID := updatePost.ID

	// 判断是否是作者
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("update post rejected: post not found", "id", postID)
			return errs.ErrNotFound
		}
		slog.Error("get post before update failed", "id", postID, "error", err)
		return errs.ErrInternal
	} else if post.User.UserID != userID { // 没有权利修改
		slog.Info("update post rejected: unauthenticated", "user_id", userID, "post_id", postID)
		return errs.ErrUnauthenticated
	}

	if err := svc.postRepo.Update(ctx, updatePost); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("update post skipped: post not found", "id", postID)
			return errs.ErrNotFound
		}
		slog.Error("update post failed", "id", postID, "user_id", userID, "error", err)
		return errs.ErrInternal
	}

	return nil
}

// ListByPage 按页获取帖子列表
func (svc *postService) ListByPage(ctx context.Context, pageNo int, pageSize int) (int64, []domain.Post, error) {
	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPage(ctx, pageNo, pageSize)
	if err != nil {
		slog.Error("get post by page", "error", err, "page_no", pageNo, "page_size", pageSize)
		return 0, nil, errs.ErrNotFound
	}

	for i, post := range posts {
		// 查找当前帖子的 Tag
		tags, _ := svc.FindTagsByPostID(ctx, post.ID)
		posts[i].Tags = tags
	}
	return total, posts, nil
}

// ListByPageAndUid 根据作者 ID 获取帖子简要信息列表
func (svc *postService) ListByPageAndUid(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []domain.Post, error) {
	total, posts, err := svc.postRepo.GetByAuthor(ctx, userID, pageNo, pageSize)
	if err != nil {
		slog.Error("get post by page and uid failed", "error", err, "user_id", userID, "page_no", pageNo, "page_size", pageSize)
		return 0, nil, errs.ErrNotFound
	}

	return total, posts, nil
}

// ListByPageAndTag 根据 Tag 分页查找帖子
func (svc *postService) ListByPageAndTag(ctx context.Context, tagName string, pageNo int, pageSize int) (int64, []domain.Post, error) {
	tag, err := svc.tagRepo.GetByName(ctx, tagName)
	if err != nil {
		slog.Info("list posts by tag skipped: tag not found", "tag", tagName)
		return 0, nil, errs.ErrNotFound
	}

	// todo 避免性能问题，优化 SQL
	// 获取帖子总数和当前页帖子列表
	total, posts, err := svc.postRepo.GetByPageAndTag(ctx, tag.ID, pageNo, pageSize)
	if err != nil {
		slog.Info("list posts by tag skipped: posts not found", "tag", tagName, "page_no", pageNo, "page_size", pageSize)
		return 0, nil, errs.ErrNotFound
	}

	for i, post := range posts {
		// 查找当前帖子的 Tag
		tags, _ := svc.FindTagsByPostID(ctx, post.ID)
		posts[i].Tags = tags
	}
	return total, posts, nil
}

// Belong 判断登录用户是否是帖子作者
func (svc *postService) Belong(ctx context.Context, userID int64, postID int64) (bool, error) {
	// 查找帖子详情
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("check post owner skipped: post not found", "id", postID)
			return false, errs.ErrNotFound
		}
		slog.Error("get post before owner check failed", "id", postID, "error", err)
		return false, errs.ErrInternal
	}

	if post.User.UserID != userID {
		return false, nil
	}
	return true, nil
}

// Delete 删除帖子
func (svc *postService) Delete(ctx context.Context, userID int64, postID int64) error {
	// 判断登录用户是否是作者
	post, err := svc.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Info("delete post skipped: post not found", "id", postID)
			return errs.ErrNotFound
		}
		slog.Error("get post before delete failed", "id", postID, "error", err)
		return errs.ErrInternal
	} else if post.User.UserID != userID {
		slog.Info("delete post rejected: unauthenticated", "user_id", userID, "post_id", postID)
		return errs.ErrUnauthenticated
	}

	// 删除帖子
	if err := svc.postRepo.Delete(ctx, postID, userID); err != nil {
		slog.Error("delete post failed", "id", post.ID, "user_id", userID, "error", err)
		return errs.ErrInternal
	}
	return nil
}

// BindTag 将 Tag 绑定到 Post
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
			if errors.Is(err, repository.ErrUniqueKey) {
				slog.Info("bind post tag skipped: already bound", "post_id", postTag.PostID, "tag_id", postTag.TagID)
				return errs.ErrAlreadyExits
			}
			slog.Error("bind post tag failed", "post_id", postTag.PostID, "tag_id", postTag.TagID, "error", err)
			return errs.ErrInternal
		}
	}

	return nil
}

// FindTagsByPostID 根据帖子 ID 查找 Tag
func (svc *postService) FindTagsByPostID(ctx context.Context, pid int64) ([]string, error) {
	var empty []string
	res, err := svc.tagRepo.FindTagsByPostID(ctx, pid)
	if err != nil {
		slog.Error("find post tags failed", "post_id", pid, "error", err)
		return empty, errs.ErrInternal
	}
	return res, nil
}
