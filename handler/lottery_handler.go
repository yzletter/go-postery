package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
)

type LotteryHandler struct {
	lotterySvc service.LotteryService
}

func NewLotteryHandler(lotterySvc service.LotteryService) *LotteryHandler {
	return &LotteryHandler{
		lotterySvc: lotterySvc,
	}
}

func (hdl *LotteryHandler) GetAllGifts(ctx *gin.Context) {

}

func (hdl *LotteryHandler) Lottery(ctx *gin.Context) {

}
func (hdl *LotteryHandler) GiveUp(ctx *gin.Context) {

}
func (hdl *LotteryHandler) Pay(ctx *gin.Context) {

}
func (hdl *LotteryHandler) Result(ctx *gin.Context) {

}
