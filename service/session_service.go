package service

import (
	"context"

	sessiondto "github.com/yzletter/go-postery/dto/session"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/model"
	"github.com/yzletter/go-postery/repository"
	"github.com/yzletter/go-postery/service/ports"
)

type sessionService struct {
	sessionRepo repository.SessionRepository
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	mq          ports.SessionMQ // Session 需要用到的 MQ 接口
}

func NewSessionService(sessionRepo repository.SessionRepository, messageRepo repository.MessageRepository, userRepo repository.UserRepository, mq ports.SessionMQ) SessionService {
	return &sessionService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		userRepo:    userRepo,
		mq:          mq,
	}
}

func (svc *sessionService) ListByUid(ctx context.Context, uid int64) ([]sessiondto.DTO, error) {
	var empty []sessiondto.DTO
	sessions, err := svc.sessionRepo.ListByUid(ctx, uid)
	if err != nil {
		return empty, errno.ErrServerInternal
	}

	var sessionDTOs []sessiondto.DTO
	for _, session := range sessions {
		// 获取对方字段
		if session.TargetType == 1 {
			// 私聊
			targetUser, err := svc.userRepo.GetByID(ctx, session.TargetID)
			if err != nil {
				targetUser = &model.User{}
			}
			sessionDTO := sessiondto.ToDTO(session, targetUser)
			sessionDTOs = append(sessionDTOs, sessionDTO)
		} else {
			// todo 群聊查 Group 表
		}
	}

	return sessionDTOs, nil
}
