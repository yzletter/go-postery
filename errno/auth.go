package errno

import "errors"

var (
	ErrJwtInvalidParam       = errors.New("jwt 传入非法参数")
	ErrJwtMarshalFailed      = errors.New("jwt json 序列化失败")
	ErrJwtBase64DecodeFailed = errors.New("jwt json base64 解码失败")
	ErrJwtUnMarshalFailed    = errors.New("jwt json 反序列化失败")
	ErrJwtInvalidTime        = errors.New("jwt 时间错误")
	ErrJwtTokenIssueFailed   = &Error{40009, 409, "JwtToken 签发失败"}
	ErrJwtTokenVerifyFailed  = &Error{40009, 409, "JwtToken 校验失败"}
	ErrInvalidTokenType      = &Error{40009, 409, "JwtToken 类别错误"}
	ErrLogoutFailed          = &Error{40009, 409, "退出登录失败"}
)
