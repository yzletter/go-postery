package service

import (
	"errors"

	"github.com/yzletter/go-postery/repository"
)

var (
	ErrJwtInvalidParam       = errors.New("jwt 传入非法参数")
	ErrJwtMarshalFailed      = errors.New("jwt json 序列化失败")
	ErrJwtBase64DecodeFailed = errors.New("jwt json base64 解码失败")
	ErrJwtUnMarshalFailed    = errors.New("jwt json 反序列化失败")
	ErrJwtInvalidTime        = errors.New("jwt 时间错误")
)

var (
	ErrInvalidParam  = errors.New("参数错误")
	ErrEncryptFailed = errors.New("加密失败")
	ErrDuplicated    = errors.New("用户名或邮箱重复")

	ErrPasswordWeak   = errors.New("密码强度太弱")
	ErrServerInternal = errors.New("内部错误")
	ErrNotFound       = errors.New("资源不存在")
	ErrPassword       = errors.New("密码错误")
)

func toServiceErr(err error) error {
	switch {
	case errors.Is(err, repository.ErrServerInternal):
		return ErrServerInternal
	case errors.Is(err, repository.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, repository.ErrConflict):
		return ErrDuplicated
	default:
		return ErrServerInternal
	}
}
