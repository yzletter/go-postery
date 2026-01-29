package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/auth/conf"
	"github.com/yzletter/go-postery/bff_/utils"
	"github.com/yzletter/go-postery/bff_/utils/response"
	"github.com/yzletter/go-postery/errno"
)

type FollowHandler struct {
	followSvc service.FollowService
	userSvc   service.UserService
}

func NewFollowHandler(followSvc service.FollowService, userSvc service.UserService) *FollowHandler {
	return &FollowHandler{
		followSvc: followSvc,
		userSvc:   userSvc,
	}
}
