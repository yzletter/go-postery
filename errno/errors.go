package errno

type Error struct {
	Code       int
	HTTPStatus int
	Msg        string
}

func (e *Error) Error() string { return e.Msg }

var (
	ErrServerInternal = &Error{50001, 500, "系统繁忙，请稍后重试"}
	ErrInvalidParam   = &Error{40001, 400, "参数错误"}
)
var (
	ErrUserNotFound      = &Error{40004, 404, "用户不存在"}
	ErrUserDuplicated    = &Error{40009, 409, "用户名或邮箱已存在"}
	ErrPasswordWeak      = &Error{40009, 409, "密码强度过低"}
	ErrInvalidCredential = &Error{40009, 409, "账号或密码错误"}
	ErrUserNotLogin      = &Error{40003, 401, "用户未登录"}
	ErrUnauthorized      = &Error{40003, 401, "没有权限"}
	ErrConflict          = &Error{40009, 409, "资源冲突"}
)
var (
	ErrLogoutFailed = &Error{40009, 409, "登出失败"}
)

var (
	ErrPostNotFound     = &Error{40004, 404, "帖子不存在"}
	ErrDuplicatedLike   = &Error{40004, 404, "重复点赞帖子"}
	ErrDuplicatedUnLike = &Error{40004, 404, "重复取消点赞帖子"}
)

var (
	ErrCommentNotFound = &Error{40004, 404, "评论不存在"}
)
