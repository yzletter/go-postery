package errno

type Error struct {
	Code       int
	HTTPStatus int
	Msg        string
}

func (e *Error) Error() string { return e.Msg }

// 通用错误 Code 1000x
var (
	ErrServerInternal = &Error{10001, 500, "系统繁忙，请稍后重试"}
	ErrInvalidParam   = &Error{10002, 400, "参数错误"}
)

// User 错误 Code 2000X
var (
	ErrUserNotFound       = &Error{20001, 404, "用户不存在"}
	ErrUserDuplicated     = &Error{20002, 409, "用户名或邮箱已存在"}
	ErrPasswordWeak       = &Error{20003, 400, "密码强度过低"}
	ErrInvalidCredential  = &Error{20004, 401, "账号或密码错误"}
	ErrUserNotLogin       = &Error{20005, 401, "用户未登录"}
	ErrUnauthorized       = &Error{20006, 403, "没有权限"}
	ErrLogoutFailed       = &Error{20007, 500, "登出失败"}
	ErrOldPasswordInvalid = &Error{20008, 401, "旧密码错误"}
)

// Post 错误 Code 3000X
var (
	ErrPostNotFound     = &Error{30001, 404, "帖子不存在"}
	ErrDuplicatedLike   = &Error{30002, 409, "已经点赞过该帖子"}
	ErrDuplicatedUnLike = &Error{30003, 409, "尚未点赞，无法取消"}
)

// Comment 错误 Code 4000X

var (
	ErrCommentNotFound = &Error{40001, 404, "评论不存在"}
)

// Tag 错误 Code 5000X

var (
	ErrTagDuplicatedBind = &Error{50001, 409, "标签重复绑定"}
)

// Follow 错误 Code 6000X
var (
	ErrDuplicatedFollow   = &Error{60001, 409, "已经关注过该用户"}
	ErrDuplicatedUnFollow = &Error{60002, 409, "尚未关注，无法取消"}
)

var (
	ErrInvalidSMSCode = &Error{70001, 401, "验证码验证失败"}
)
