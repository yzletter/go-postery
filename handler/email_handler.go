package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/dto/email"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils/response"
)

type EmailHandler struct {
	emailSvc service.EmailService
}

func NewEmailHandler(emailSvc service.EmailService) *EmailHandler {
	return &EmailHandler{emailSvc: emailSvc}
}

func (hdl *EmailHandler) Send(ctx *gin.Context) {
	// 获取参数
	var req email.SendEmailCodeRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// todo 参数校验

	// 发送邮件
	err = hdl.emailSvc.SendSMS(ctx, req.EmailAddress)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "发送邮箱验证码成功", nil)
}
