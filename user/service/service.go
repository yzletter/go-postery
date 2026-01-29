package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bytedance/sonic"
	"github.com/segmentio/kafka-go"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/user/dto"
	"github.com/yzletter/go-postery/user/errs"
	"github.com/yzletter/go-postery/user/model"
	"github.com/yzletter/go-postery/user/repository"
	"github.com/yzletter/go-postery/user/service/ports"
)

type userService struct {
	userRepository   repository.UserRepository   // 依赖 UserRepository
	followRepository repository.FollowRepository // 依赖 FollowRepository
	kafkaConsumer    *kafka.Reader
	idGen            ports.IDGenerator
	user_grpc.UnimplementedUserServiceServer
}

func NewUserService(userRepository repository.UserRepository, followRepository repository.FollowRepository, kafkaConsumer *kafka.Reader, idGen ports.IDGenerator) UserService {
	return &userService{
		userRepository:                 userRepository,
		followRepository:               followRepository,
		kafkaConsumer:                  kafkaConsumer,
		idGen:                          idGen,
		UnimplementedUserServiceServer: user_grpc.UnimplementedUserServiceServer{},
	}
}

func (svc *userService) GetProfileById(ctx context.Context, req *user_grpc.GetProfileByIdRequest) (*user_grpc.UserDetail, error) {
	empty := new(user_grpc.UserDetail)

	if req.ID <= 0 {
		slog.Error("Invalid User ID")
		return empty, errs.ErrInvalidArgument
	}

	// 获取用户
	profile, err := svc.userRepository.GetProfileByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("User Not Found", "error", err)
			return empty, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return empty, errs.ErrInternal
	}

	if profile == nil {
		slog.Error("User Not Found", "error", err)
		return empty, errs.ErrNotFound
	}

	return dto.ToUserDetail(profile), nil
}

func (svc *userService) UpdateProfile(ctx context.Context, req *user_grpc.UpdateProfileRequest) (*user_grpc.UpdateProfileResponse, error) {
	empty := new(user_grpc.UpdateProfileResponse)

	if req.ID <= 0 {
		slog.Error("Invalid User ID")
		return empty, errs.ErrInvalidArgument
	}

	// 将 Request 转为 Model, 主要是 Birthday 从 RFC3339 string 转为 Time.time
	profile := dto.UpdateProfileRequestToModel(req)

	updates := map[string]any{
		"nickname": profile.Nickname,
		"avatar":   profile.Avatar,
		"bio":      profile.Bio,
		"gender":   profile.Gender,
		"birthday": profile.Birthday,
		"location": profile.Location,
		"country":  profile.Country,
	}

	if err := svc.userRepository.UpdateProfile(ctx, req.ID, updates); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("User Not Found", "error", err)
			return empty, errs.ErrNotFound
		}
		slog.Error("Server Internal Error", "error", err)
		return empty, errs.ErrInternal
	}

	return &user_grpc.UpdateProfileResponse{}, nil
}

func (svc *userService) Top(ctx context.Context, req *user_grpc.TopRequest) (*user_grpc.TopResponse, error) {
	empty := new(user_grpc.TopResponse)
	profiles, scores, err := svc.userRepository.Top(ctx)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return empty, errs.ErrInternal
	}

	var topUsers []*user_grpc.TopUser
	for idx, profile := range profiles {
		topUsers = append(topUsers, dto.ToTopUser(profile, scores[idx]))
	}

	return &user_grpc.TopResponse{TopUsers: topUsers}, nil
}

// Follow 关注用户
func (svc *userService) Follow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error) {
	res, err := svc.followRepository.Exists(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return &user_grpc.FollowEmptyResponse{}, errs.ErrInternal
	}

	if res == 1 || res == 3 { // 已经关注过了
		slog.Error("Duplicated Follow")
		return &user_grpc.FollowEmptyResponse{}, errs.ErrAlreadyExits
	}

	follow := &model.Follow{
		ID:         svc.idGen.NextID(),
		FollowerID: req.FollowerID,
		FolloweeID: req.FolloweeID,
	}
	if err = svc.followRepository.Create(ctx, follow); err != nil {
		if errors.Is(err, repository.ErrUniqueKey) { // 检查过仍冲突
			slog.Error("Duplicated Follow")
			return &user_grpc.FollowEmptyResponse{}, errs.ErrAlreadyExits
		}
		slog.Error("Server Internal Error", "error", err)
		return &user_grpc.FollowEmptyResponse{}, errs.ErrInternal
	}
	_ = svc.userRepository.ChangeScore(ctx, req.FolloweeID, 1)
	return &user_grpc.FollowEmptyResponse{}, nil
}

// UnFollow 取消关注
func (svc *userService) UnFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.FollowEmptyResponse, error) {
	res, err := svc.followRepository.Exists(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return &user_grpc.FollowEmptyResponse{}, errs.ErrInternal
	}

	if res == 2 || res == 0 { // 只有对方关注了我，或者互不关注
		slog.Error("Duplicated UnFollow")
		return &user_grpc.FollowEmptyResponse{}, errs.ErrAlreadyExits
	}

	if err := svc.followRepository.Delete(ctx, req.FollowerID, req.FolloweeID); err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			slog.Error("Duplicated UnFollow")
			return &user_grpc.FollowEmptyResponse{}, errs.ErrAlreadyExits
		}
		slog.Error("Server Internal Error", "error", err)
		return &user_grpc.FollowEmptyResponse{}, errs.ErrInternal
	}

	_ = svc.userRepository.ChangeScore(ctx, req.FolloweeID, -1)

	return &user_grpc.FollowEmptyResponse{}, nil
}

// IfFollow 判断关注关系
func (svc *userService) IfFollow(ctx context.Context, req *user_grpc.FollowCommonRequest) (*user_grpc.IfFollowResponse, error) {
	res, err := svc.followRepository.Exists(ctx, req.FollowerID, req.FolloweeID)
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return &user_grpc.IfFollowResponse{Result: -1}, errs.ErrInternal // 数据库内部错误
	}
	return &user_grpc.IfFollowResponse{Result: int32(res)}, nil
}

// ListFollowersByPage 按页查找粉丝
func (svc *userService) ListFollowersByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	var empty = new(user_grpc.ListFollowResponse)
	total, followersId, err := svc.followRepository.GetFollowers(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return empty, errs.ErrInternal
	}

	userBriefs := make([]*user_grpc.UserBrief, 0)
	for _, id := range followersId {
		profile, err := svc.userRepository.GetProfileByID(ctx, id)
		if err != nil {
			continue
		}
		userBrief := dto.ToUserBrief(profile)
		userBriefs = append(userBriefs, userBrief)
	}

	return &user_grpc.ListFollowResponse{Count: uint64(total), UserBriefs: userBriefs}, nil

}

// ListFolloweesByPage 按页查找关注的人
func (svc *userService) ListFolloweesByPage(ctx context.Context, req *user_grpc.ListFollowRequest) (*user_grpc.ListFollowResponse, error) {
	var empty = new(user_grpc.ListFollowResponse)
	total, followeesId, err := svc.followRepository.GetFollowees(ctx, req.UserID, int(req.PageNo), int(req.PageSize))
	if err != nil {
		slog.Error("Server Internal Error", "error", err)
		return empty, errs.ErrInternal
	}

	userBriefs := make([]*user_grpc.UserBrief, 0)
	for _, id := range followeesId {
		profile, err := svc.userRepository.GetProfileByID(ctx, id)
		if err != nil {
			continue
		}
		userBrief := dto.ToUserBrief(profile)
		userBriefs = append(userBriefs, userBrief)
	}

	return &user_grpc.ListFollowResponse{Count: uint64(total), UserBriefs: userBriefs}, nil
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
			var payload model.InitUserScoreEventPayload
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
