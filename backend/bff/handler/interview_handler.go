package handler

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	interview_grpc "github.com/yzletter/go-postery/api/proto/interview/v1"
	interview_dto "github.com/yzletter/go-postery/backend/bff/dto/interview"
	"github.com/yzletter/go-postery/backend/bff/errno"
	"github.com/yzletter/go-postery/backend/conf"
	"github.com/yzletter/go-postery/backend/grpc/manager"
	"github.com/yzletter/go-postery/backend/utils"
	"github.com/yzletter/go-postery/backend/utils/response"
	"google.golang.org/grpc/codes"
)

type InterviewHandler struct {
	interviewSvc manager.InterviewClient
}

// NewInterviewHandler 构造函数
func NewInterviewHandler(interviewSvc manager.InterviewClient) *InterviewHandler {
	return &InterviewHandler{
		interviewSvc: interviewSvc,
	}
}

func (hdl *InterviewHandler) RegisterRouter(engine *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	interviews := engine.Group("/interviews")
	interviews.POST("/questions/callback", hdl.UploadQuestionsCallback) // POST /interviews/questions/callback OSS 题库上传回调

	authedInterviews := interviews.Group("")
	authedInterviews.Use(authMiddleware)
	authedInterviews.GET("/questions/upload", hdl.UploadQuestionsSign) // GET /interviews/questions/upload 获取题库上传签名
}

// UploadQuestionsSign 获取上传题库 OSS 签名
func (hdl *InterviewHandler) UploadQuestionsSign(ctx *gin.Context) {
	// 由于前面有 Auth 中间件, 能走到这里默认上下文里已经被 Auth 塞了 uid, 直接拿即可
	uid, err := utils.GetUidFromCTX(ctx, conf.UserIDInContext)
	if err != nil {
		response.Error(ctx, errno.ErrUserNotLogin)
		return
	}

	resp, err := hdl.interviewSvc.UploadQuestionsSign(ctx, &interview_grpc.UploadQuestionsSignRequest{UserID: uid})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
		}, errno.ErrServerInternal), nil)
		return
	}

	response.Success(ctx, "获取签名成功", interview_dto.OSSSignDTO{Response: resp.Response})
}

// UploadQuestionsCallback 处理 OSS 题库上传回调
func (hdl *InterviewHandler) UploadQuestionsCallback(ctx *gin.Context) {
	// 对请求进行验签
	if ok, err := utils.VerifyOSS(ctx.Request); !ok || err != nil {
		slog.Warn("oss callback signature invalid", "error", err, "method", ctx.Request.Method, "path", ctx.Request.URL.Path, "client_ip", ctx.ClientIP())
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// Body 绑定结构体
	var req interview_dto.UploadCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		slog.Warn("bind request failed", "error", utils.BindErrMsg(err))
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 验证 Bucket
	if req.Bucket != "go-postery" {
		slog.Warn("invalid oss bucket", "bucket", req.Bucket)
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	// 获取 uid 和 objectName
	// interviews/questions/uid/filename
	segments := strings.Split(req.Object, "/")
	if len(segments) != 4 || segments[0] != "interviews" || segments[1] != "questions" || segments[3] == "" {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	uid, err := strconv.ParseInt(segments[2], 10, 64)
	if err != nil || uid <= 0 {
		response.Error(ctx, errno.ErrInvalidParam)
		return
	}

	_, err = hdl.interviewSvc.UploadQuestionsCallback(ctx, &interview_grpc.UploadQuestionsCallbackRequest{
		UserID:     uid,
		ObjectName: req.Object,
	})
	if err != nil {
		response.Error(ctx, mapGRPCErr(err, map[codes.Code]*errno.Error{
			codes.InvalidArgument: errno.ErrInvalidParam,
		}, errno.ErrServerInternal), nil)
		return
	}

	response.Success(ctx, "回调成功", nil)
}
