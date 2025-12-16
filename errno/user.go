package errno

var (
	ErrInvalidParam      = &Error{40001, 400, "参数错误"}
	ErrUserNotFound      = &Error{40004, 404, "用户不存在"}
	ErrUserDuplicated    = &Error{40009, 409, "资源冲突"}
	ErrPasswordWeak      = &Error{40009, 409, "密码强度过低"}
	ErrInvalidCredential = &Error{40009, 409, "账号或密码错误"}
	ErrNotLogin          = &Error{40003, 401, "未登录"}
	ErrUnauthorized      = &Error{40003, 401, "无权限"}
	ErrConflict          = &Error{40009, 409, "资源冲突"}
)
