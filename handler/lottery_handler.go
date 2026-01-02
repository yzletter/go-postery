package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils/response"
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
	giftDTOs, err := hdl.lotterySvc.GetAllGifts(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取全部奖品成功", giftDTOs)
}

func (hdl *LotteryHandler) Lottery(ctx *gin.Context) {

}
func (hdl *LotteryHandler) GiveUp(ctx *gin.Context) {

}
func (hdl *LotteryHandler) Pay(ctx *gin.Context) {

}
func (hdl *LotteryHandler) Result(ctx *gin.Context) {

}
