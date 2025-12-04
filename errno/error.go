package errno

import "errors"

var (
	ErrRecordNotFound = errors.New("数据库记录未找到")
	ErrCreateFailed   = errors.New("数据库创建失败")
	ErrDeleteFailed   = errors.New("数据库删除失败")
	ErrGetFailed      = errors.New("数据库查询失败")
)
