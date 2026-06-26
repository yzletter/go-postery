package errno

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	Code       int
	HTTPStatus int
	Msg        string
}

func (e *Error) Error() string { return e.Msg }

// 通用错误 Code 1000x
var (
	ErrServerInternal     = &Error{10001, 500, "系统繁忙，请稍后重试"}
	ErrInvalidParam       = &Error{10002, 400, "参数错误"}
	ErrServiceUnavailable = &Error{10003, 503, "下游服务暂时不可用，请稍后重试"}
)

// Auth
var (
	ErrInvalidCode        = &Error{70001, 401, "验证码验证失败"}
	ErrSendToFrequent     = &Error{70002, 401, "验证码发送过于频繁"}
	ErrCodeNotFound       = &Error{70003, 401, "验证码不存在"}
	ErrEmailCodeInvalid   = &Error{70004, 401, "邮箱或验证码错误"}
	ErrPhoneCodeInvalid   = &Error{70005, 401, "手机号或验证码错误"}
	ErrSamePassword       = &Error{70006, 401, "两次密码一致"}
	ErrPasswordWeak       = &Error{70007, 400, "密码强度过低"}
	ErrInvalidCredential  = &Error{70008, 401, "账号或密码错误"}
	ErrUnauthorized       = &Error{70009, 403, "没有权限"}
	ErrLogoutFailed       = &Error{70010, 500, "登出失败"}
	ErrOldPasswordInvalid = &Error{70011, 401, "旧密码错误"}
	ErrSetPassword        = &Error{70012, 401, "初始化密码失败"}
	ErrUserNotLogin       = &Error{70013, 401, "用户未登录"}
	ErrNotSetPassword     = &Error{70014, 401, "未初始化密码"}
	ErrPhoneNotBound      = &Error{70015, 400, "请先绑定手机号"}
)

// User 错误 Code 2000X
var (
	ErrUserNotFound       = &Error{20001, 404, "用户不存在"}
	ErrUserDuplicated     = &Error{20002, 409, "用户已存在"}
	ErrDuplicatedFollow   = &Error{20003, 409, "已经关注过该用户"}
	ErrDuplicatedUnFollow = &Error{20004, 409, "尚未关注，无法取消"}
	ErrFollowYourself     = &Error{20005, 409, "无法对自己进行关注操作"}
	ErrNicknameDuplicated = &Error{20006, 409, "昵称已存在"}
)

// Post 错误 Code 3000X
var (
	ErrPostNotFound      = &Error{30001, 404, "帖子不存在"}
	ErrDuplicatedLike    = &Error{30002, 409, "已经点赞过该帖子"}
	ErrDuplicatedUnLike  = &Error{30003, 409, "尚未点赞，无法取消"}
	ErrCommentNotFound   = &Error{30004, 404, "评论不存在"}
	ErrTagDuplicatedBind = &Error{30005, 409, "标签重复绑定"}
)

// Lottery 错误
var (
	ErrGiftNotFound  = &Error{40001, 404, "奖品不存在"}
	ErrNotLottery    = &Error{40002, 404, "没有抢到该商品，或支付时限已过"}
	ErrOrderNotFound = &Error{40003, 404, "订单不存在"}
	ErrLotteryNoting = &Error{40004, 404, "没有抽到奖品"}
)

func GetGRPCErrCode(err error) codes.Code {
	st, ok := status.FromError(err)
	if !ok {
		// 非 gRPC 错误（网络、panic、ctx）
		return codes.Unknown
	}

	return st.Code()
}
