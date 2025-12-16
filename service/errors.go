package service

import "errors"

var (
	ErrJwtInvalidParam       = errors.New("jwt 传入非法参数")
	ErrJwtMarshalFailed      = errors.New("jwt json 序列化失败")
	ErrJwtBase64DecodeFailed = errors.New("jwt json base64 解码失败")
	ErrJwtUnMarshalFailed    = errors.New("jwt json 反序列化失败")
	ErrJwtInvalidTime        = errors.New("jwt 时间错误")
)

var (
	ErrInvalidParam   = errors.New("参数错误")
	ErrEncryptFailed  = errors.New("加密失败")
	ErrServerInternal = errors.New("注册失败, 请稍后重试")
	ErrDuplicated     = errors.New("用户名或邮箱重复")
	ErrPasswordWeak   = errors.New("密码强度太弱")
)
