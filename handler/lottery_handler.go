package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	orderdto "github.com/yzletter/go-postery/dto/order"
	"github.com/yzletter/go-postery/errno"
	"github.com/yzletter/go-postery/service"
	"github.com/yzletter/go-postery/utils"
	"github.com/yzletter/go-postery/utils/response"
)

type LotteryHandler struct {
	lotterySvc service.LotteryService
}

func NewLotteryHandler(lotterySvc service.LotteryService) *LotteryHandler {
	ctx := context.Background()

	// 开启消费协程
	go lotterySvc.Consume(ctx)

	// 初始化缓存库存
	lotterySvc.InitCacheInventory(ctx)

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
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	giftDTO, err := hdl.lotterySvc.Lottery(ctx, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "抽奖成功", giftDTO)
}

func (hdl *LotteryHandler) GiveUp(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	var giveReq orderdto.GiveUpRequest
	err = ctx.ShouldBindJSON(&giveReq)
	if err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 登录用户与放弃用户不一致
	if uid != giveReq.UserID {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	err = hdl.lotterySvc.GiveUp(ctx, giveReq.UserID, giveReq.GiftID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "放弃支付成功", nil)
}

func (hdl *LotteryHandler) Pay(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	var payReq orderdto.PayRequest
	err = ctx.ShouldBindJSON(&payReq)
	if err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	fmt.Println(payReq)

	// 登录用户与支付用户不一致
	if uid != payReq.UserID {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	// 进行支付
	err = hdl.lotterySvc.Pay(ctx, payReq.UserID, payReq.GiftID)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "支付成功", nil)
}

func (hdl *LotteryHandler) Result(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	// 获取结果
	orderDTO, err := hdl.lotterySvc.Result(ctx, uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, "获取结果成功", orderDTO)
}
