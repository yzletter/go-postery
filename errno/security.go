package errno

var (
	ErrPasswordEncryptFailed     = &Error{Code: 0, HTTPStatus: 0, Msg: "密码加密失败"}
	ErrMismatchedHashAndPassword = &Error{Code: 0, HTTPStatus: 0, Msg: "密码校验不正确"}
)
