package client

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	gobreaker "github.com/sony/gobreaker/v2"
	"github.com/yzletter/go-postery/microservice-backend/bff/errno"
	_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var defaultBreakerPool = NewBreaker()

// BreakerPool 熔断器
type BreakerPool struct {
	mu sync.RWMutex
	mp map[string]*gobreaker.CircuitBreaker[any] // 对应每个下游的 Breaker
}

func NewBreaker() *BreakerPool {
	return &BreakerPool{mp: make(map[string]*gobreaker.CircuitBreaker[any])}
}

// GetBreaker 获取当前下游的 Breaker
func (pool *BreakerPool) GetBreaker(key string) *gobreaker.CircuitBreaker[any] {
	// 上读锁
	pool.mu.RLock()
	breaker, exist := pool.mp[key]
	pool.mu.RUnlock()

	// 存在 breaker 直接返回
	if exist {
		return breaker
	}

	// 不存在 breaker 进行创建
	pool.mu.Lock()
	defer pool.mu.Unlock() // defer 解锁

	// double check
	if breaker, exist := pool.mp[key]; exist {
		return breaker
	}

	// 开始创建 Breaker

	// 相关配置
	setting := gobreaker.Settings{
		Name:         key,              // 熔断器名字 服务名 + 方法名
		MaxRequests:  3,                // Half-Open 状态下，最多允许多少个探测请求通过
		Interval:     30 * time.Second, // 统计窗口总长度
		BucketPeriod: 3 * time.Second,  // 每个小窗口大小
		Timeout:      15 * time.Second, // 熔断器从 Open 状态切到 Half-Open 之前等待多久
		// ReadyToTrip 什么时候打开熔断器
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			validRequests := counts.Requests
			if counts.TotalExclusions < validRequests {
				validRequests -= counts.TotalExclusions
			} else {
				validRequests = 0
			}

			// 样本过少，放行
			if validRequests < 20 {
				return false
			}

			// 计算失败率：总失败数 / 有效请求数
			failRate := float64(counts.TotalFailures) / float64(validRequests)

			// 满足下面任一条件就打开熔断器：
			// 1. 失败率 >= 50%
			// 2. 连续失败 >= 5 次
			return failRate >= 0.5 || counts.ConsecutiveFailures >= 5
		},
		// 回调函数
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			slog.Info("Breaker State Changed", "name", name, "from", from.String(), "to", to.String())
		},
		// 定义什么样的请求算成功
		IsSuccessful: func(err error) bool {
			// 没有错误，当然是成功
			if err == nil {
				return true
			}

			// 解析 gRPC status
			st, ok := status.FromError(err)
			if !ok {
				// 不是标准 gRPC 错误，按失败处理，采用保守策略
				return false
			}

			switch st.Code() {
			// 这些通常是“业务错误”或“请求本身有问题”，不代表下游服务挂了，所以这里按“成功”对待
			// 意思是：不要把 breaker 打开。
			case codes.InvalidArgument,
				codes.NotFound,
				codes.AlreadyExists,
				codes.FailedPrecondition,
				codes.PermissionDenied,
				codes.Unauthenticated:
				return true

			// 这些更像“依赖不健康”或“资源不足”，应该计入失败，从而参与熔断判断
			case codes.DeadlineExceeded,
				codes.Unavailable,
				codes.Internal,
				codes.ResourceExhausted,
				codes.Unknown:
				return false

			// 其他默认按失败处理，采用保守策略
			default:
				return false
			}
		},
		// IsExcluded 用来排除“不参与熔断统计”的错误
		IsExcluded: func(err error) bool {
			// 调用方主动取消请求，一般不认为是下游故障，所以排除。
			if errors.Is(err, context.Canceled) {
				return true
			}
			// gRPC 场景下，context.Canceled 也可能被包装成 codes.Canceled
			st, ok := status.FromError(err)
			return ok && st.Code() == codes.Canceled
		},
	}

	// 创建 Breaker
	breaker = gobreaker.NewCircuitBreaker[any](setting)

	// 放入缓存池
	pool.mp[key] = breaker
	return breaker
}

// CircuitBreakerDialOption 返回带统一 breaker 配置的 gRPC client 选项。
func CircuitBreakerDialOption() _grpc.DialOption {
	return _grpc.WithChainUnaryInterceptor(UnaryClientCircuitBreaker(defaultBreakerPool))
}

// UnaryClientCircuitBreaker 返回 grpc 拦截器
func UnaryClientCircuitBreaker(pool *BreakerPool) _grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *_grpc.ClientConn, invoker _grpc.UnaryInvoker, opts ..._grpc.CallOption) error {
		// 服务名 + 方法
		key := cc.Target() + ":" + method

		// 拿到下游的 Breaker
		breaker := pool.GetBreaker(key)

		// 通过 breaker.Execute 包裹真正的 RPC 调用
		// 作用：
		// 1. breaker 关闭时，正常放行
		// 2. breaker 打开时，直接拒绝，不再发请求
		// 3. breaker 半开时，只放少量探测请求
		_, err := breaker.Execute(func() (any, error) {
			// invoker 才是真正的 gRPC 调用动作
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})

		switch {
		case errors.Is(err, gobreaker.ErrOpenState):
			return errno.ErrUnavailable
		case errors.Is(err, gobreaker.ErrTooManyRequests):
			return errno.ErrUnavailable
		default:
			return err
		}
	}
}
