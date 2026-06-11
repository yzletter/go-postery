package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	lottery_grpc "github.com/yzletter/go-postery/api/proto/lottery/v1"
	user_grpc "github.com/yzletter/go-postery/api/proto/user/v1"
	"github.com/yzletter/go-postery/backend/conf"
	grpcclient "github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/micro/bff/dto/lottery"
	"github.com/yzletter/go-postery/backend/micro/bff/errno"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type LotteryHandler struct {
	lotterySvc grpcclient.LotteryClient
	userSvc    grpcclient.UserClient
}

func NewLotteryHandler(lotterySvc grpcclient.LotteryClient, userSvc grpcclient.UserClient) *LotteryHandler {
	return &LotteryHandler{
		lotterySvc: lotterySvc,
		userSvc:    userSvc,
	}
}

func (hdl *LotteryHandler) GetAllGifts(ctx *gin.Context) {
	resp, err := hdl.lotterySvc.GetAllGifts(ctx, &lottery_grpc.EmptyRequest{})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrGiftNotFound,
		}, errno.ErrServerInternal), []lottery.GiftDTO{})
		return
	}

	gifts := make([]lottery.GiftDTO, 0, len(resp.Gifts))
	for _, gift := range resp.Gifts {
		gifts = append(gifts, lottery.ToGiftDTO(gift))
	}

	response.Success(ctx, "获取全部奖品成功", gifts)
}

func (hdl *LotteryHandler) Lottery(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	result, err := hdl.lotterySvc.Lottery(ctx, &lottery_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, nil, errno.ErrServerInternal), lottery.LotteryResultDTO{})
		return
	}

	resultDTO := lottery.ToLotteryResultDTO(result)
	message := resultDTO.ResultDescription
	if message == "" {
		message = "抽奖成功"
	}
	response.Success(ctx, message, resultDTO)
}

func (hdl *LotteryHandler) GiveUp(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	var giveReq lottery.GiveUpRequest
	if err := ctx.ShouldBindJSON(&giveReq); err != nil {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 登录用户与放弃用户不一致
	if uid != giveReq.UserID {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	if _, err := hdl.lotterySvc.GiveUp(ctx, &lottery_grpc.LotteryCommonRequest{
		UserID:      giveReq.UserID,
		GiftID:      giveReq.GiftID,
		TempOrderID: giveReq.TempOrderID,
	}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrNotLottery,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "放弃支付成功", nil)
}

func (hdl *LotteryHandler) Pay(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	var payReq lottery.PayRequest

	if err := ctx.ShouldBindJSON(&payReq); err != nil {
		// 参数绑定失败
		slog.Error("参数绑定失败", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 登录用户与支付用户不一致
	if uid != payReq.UserID {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	// 进行支付
	if _, err := hdl.lotterySvc.Pay(ctx, &lottery_grpc.LotteryCommonRequest{
		UserID:      payReq.UserID,
		GiftID:      payReq.GiftID,
		TempOrderID: payReq.TempOrderID,
	}); err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrNotLottery,
		}, errno.ErrServerInternal), gin.H{})
		return
	}

	response.Success(ctx, "支付成功", nil)
}

func (hdl *LotteryHandler) Result(ctx *gin.Context) {
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	// 获取结果
	order, err := hdl.lotterySvc.Result(ctx, &lottery_grpc.UserID{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.NotFound: errno.ErrOrderNotFound,
		}, errno.ErrServerInternal), lottery.OrderDTO{})
		return
	}

	if order.UserID != uid {
		response.Error(ctx, errno.ErrUnauthorized)
		return
	}

	// 查询用户
	user, err := hdl.userSvc.GetProfileById(ctx, &user_grpc.GetProfileByIdRequest{ID: order.UserID})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
			codes.NotFound:        errno.ErrUserNotFound,
		}, errno.ErrServerInternal), lottery.OrderDTO{})
		return
	}

	response.Success(ctx, "获取结果成功", lottery.ToOrderDTO(order, user))
}
