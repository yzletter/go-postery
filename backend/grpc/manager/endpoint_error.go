package manager

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// isEndpointFailure 判断错误是否代表当前节点不可用, 业务错误不应该降低节点健康度
func isEndpointFailure(err error) bool {
	if err == nil {
		return false
	}

	switch status.Code(err) {
	case codes.Unavailable,
		codes.DeadlineExceeded,
		codes.ResourceExhausted,
		codes.Unknown,
		codes.Internal:
		return true
	default:
		return false
	}
}
