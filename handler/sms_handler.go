package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/sms"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils/response"
)

type SmsHandler struct {
	smsSvc service.SmsService
}

func NewSmsHandler(smsSvc service.SmsService) *SmsHandler {
	return &SmsHandler{smsSvc: smsSvc}
}

// Send 发送短信验证码
func (hdl *SmsHandler) Send(ctx *gin.Context) {
	// 获取参数
	var req sms.SendSMSCodeRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 参数校验
	if len(req.PhoneNumber) != 11 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 发送验证码
	err = hdl.smsSvc.SendSMS(ctx, req.PhoneNumber)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "短信发送成功", nil)
}
