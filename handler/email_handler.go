package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
)

type EmailHandler struct {
	emailSvc service.EmailService
}

func NewEmailHandler(emailSvc service.EmailService) *EmailHandler {
	return &EmailHandler{emailSvc: emailSvc}
}

func (hdl *EmailHandler) Send(ctx *gin.Context) {

}
