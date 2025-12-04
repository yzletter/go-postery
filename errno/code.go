package errno

const (
	CodeSuccess      = 0     // 成功
	CodeBadRequest   = 40001 // 参数错误
	CodeUnauthorized = 40003 // 未登录/无权限
	CodeServerError  = 50001 // 服务端错误
	// 其他业务错误码
)
