package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	"github.com/yzletter/go-postery/microservice-backend/user/errs"
	model2 "github.com/yzletter/go-postery/microservice-backend/user/model"
	repository2 "github.com/yzletter/go-postery/microservice-backend/user/repository"
	"github.com/yzletter/go-postery/microservice-backend/user/service/ports"
)

type userService struct {
	userRepository   repository2.UserRepository   // 依赖 UserRepository
	followRepository repository2.FollowRepository // 依赖 FollowRepository
	kafkaConsumer    *kafka.Reader
	idGen            ports.IDGenerator
}

func NewUserService(userRepository repository2.UserRepository, followRepository repository2.FollowRepository, kafkaConsumer *kafka.Reader, idGen ports.IDGenerator) UserService {
	return &userService{
		userRepository:   userRepository,
		followRepository: followRepository,
		kafkaConsumer:    kafkaConsumer,
		idGen:            idGen,
	}
}

func (svc *userService) GetProfileByID(ctx context.Context, id int64) (*model2.UserProfile, error) {
	if id <= 0 {
		slog.Error("Invalid User ID")
		return nil, errs.ErrInvalidArgument
	}

	// 获取用户
	profile, err := svc.userRepository.GetProfileByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("User Not Found", "error", err)
			return nil, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return nil, errs.ErrInternal
	}

	if profile == nil {
		slog.Error("User Not Found", "error", err)
		return nil, errs.ErrNotFound
	}

	return profile, nil
}

func (svc *userService) UpdateProfile(ctx context.Context, profile *model2.UserProfile) error {
	if profile == nil || profile.UserID <= 0 {
		slog.Error("Invalid User ID")
		return errs.ErrInvalidArgument
	}

	updates := map[string]any{
		"nickname": profile.Nickname,
		"avatar":   profile.Avatar,
		"bio":      profile.Bio,
		"gender":   profile.Gender,
		"birthday": profile.Birthday,
		"location": profile.Location,
		"country":  profile.Country,
	}

	if err := svc.userRepository.UpdateProfile(ctx, profile.UserID, updates); err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("User Not Found", "error", err)
			return errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	return nil
}

func (svc *userService) Top(ctx context.Context) ([]*model2.UserProfile, []float64, error) {
	profiles, scores, err := svc.userRepository.Top(ctx)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return nil, nil, errs.ErrInternal
	}

	return profiles, scores, nil
}

// Follow 关注用户
func (svc *userService) Follow(ctx context.Context, followerID int64, followeeID int64) error {
	res, err := svc.followRepository.Exists(ctx, followerID, followeeID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	if res == 1 || res == 3 { // 已经关注过了
		slog.Error("Duplicated Follow")
		return errs.ErrAlreadyExits
	}

	follow := &model2.Follow{
		ID:         svc.idGen.NextID(),
		FollowerID: followerID,
		FolloweeID: followeeID,
	}
	if err = svc.followRepository.Create(ctx, follow); err != nil {
		if errors.Is(err, repository2.ErrUniqueKey) { // 检查过仍冲突
			slog.Error("Duplicated Follow")
			return errs.ErrAlreadyExits
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}
	_ = svc.userRepository.ChangeScore(ctx, followeeID, 1)
	return nil
}

// UnFollow 取消关注
func (svc *userService) UnFollow(ctx context.Context, followerID int64, followeeID int64) error {
	res, err := svc.followRepository.Exists(ctx, followerID, followeeID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	if res == 2 || res == 0 { // 只有对方关注了我，或者互不关注
		slog.Error("Duplicated UnFollow")
		return errs.ErrAlreadyExits
	}

	if err := svc.followRepository.Delete(ctx, followerID, followeeID); err != nil {
		if errors.Is(err, repository2.ErrRecordNotFound) {
			slog.Error("Duplicated UnFollow")
			return errs.ErrAlreadyExits
		}
		slog.Error("Server Internal Error", "error", err)
		return errs.ErrInternal
	}

	_ = svc.userRepository.ChangeScore(ctx, followeeID, -1)

	return nil
}

// IfFollow 判断关注关系
func (svc *userService) IfFollow(ctx context.Context, followerID int64, followeeID int64) (int, error) {
	res, err := svc.followRepository.Exists(ctx, followerID, followeeID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return -1, errs.ErrInternal // 数据库内部错误
	}
	return int(res), nil
}

// ListFollowersByPage 按页查找粉丝
func (svc *userService) ListFollowersByPage(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []*model2.UserProfile, error) {
	total, followersId, err := svc.followRepository.GetFollowers(ctx, userID, pageNo, pageSize)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return 0, nil, errs.ErrInternal
	}

	userProfiles := make([]*model2.UserProfile, 0)
	for _, id := range followersId {
		profile, err := svc.userRepository.GetProfileByID(ctx, id)
		if err != nil {
			continue
		}
		userProfiles = append(userProfiles, profile)
	}

	return total, userProfiles, nil

}

// ListFolloweesByPage 按页查找关注的人
func (svc *userService) ListFolloweesByPage(ctx context.Context, userID int64, pageNo int, pageSize int) (int64, []*model2.UserProfile, error) {
	total, followeesId, err := svc.followRepository.GetFollowees(ctx, userID, pageNo, pageSize)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return 0, nil, errs.ErrInternal
	}

	userProfiles := make([]*model2.UserProfile, 0)
	for _, id := range followeesId {
		profile, err := svc.userRepository.GetProfileByID(ctx, id)
		if err != nil {
			continue
		}
		userProfiles = append(userProfiles, profile)
	}

	return total, userProfiles, nil
}

// StartInitUserScoreConsumer 开启初始化用户分数的 Kafka 消费者协程
func (svc *userService) StartInitUserScoreConsumer(ctx context.Context) {
	backoff := time.Second
	for {
		select {
		case <-ctx.Done():
			slog.Info("Close Init Score Consumer Success ...")
			return
		default:
			// Fetch 消息
			message, err := svc.kafkaConsumer.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { // 正常退出
					return
				}

				slog.Error("Fetch Message Failed", "Topic", "Session", "error", err)

				// 简单退避，避免狂刷日志
				time.Sleep(backoff)
				if backoff < 10*time.Second {
					backoff *= 2
				}
				continue
			}

			backoff = time.Second // 重置

			// 解析 JSON
			var payload model2.InitUserScoreEventPayload
			if err := sonic.Unmarshal(message.Value, &payload); err != nil {
				slog.Error("invalid message value, skip", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "value", string(message.Value), "errs", err)
				// 脏消息 Commit 掉
				_ = svc.kafkaConsumer.CommitMessages(ctx, message)
				continue
			}

			// 进行初始化用户分数
			if err := svc.userRepository.ChangeScore(ctx, payload.UserID, int(time.Now().Unix())); err != nil {
				slog.Error("Init User Score Failed", "error", err)
				time.Sleep(time.Second) // 最小退避，避免打爆
				continue                // 不 commit -> 重试
			}

			// 初始化用户分数成功, 把消息 Commit 掉
			if err := svc.kafkaConsumer.CommitMessages(ctx, message); err != nil {
				slog.Error("Commit Kafka Message Failed", "uid", payload.UserID, "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "errs", err)
				// Commit 失败通常会导致重复消费，但不会丢消息，可接受
				continue
			}
		}
	}
}
