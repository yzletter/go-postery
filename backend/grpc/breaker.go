package grpc

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	gobreaker "github.com/sony/gobreaker/v2"
	"github.com/yzletter/go-postery/backend/grpc/errs"
	_grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var defaultBreakerPool = NewBreaker()

type BreakerPool struct {
	mu sync.RWMutex
	mp map[string]*gobreaker.CircuitBreaker[any]
}

func NewBreaker() *BreakerPool {
	return &BreakerPool{mp: make(map[string]*gobreaker.CircuitBreaker[any])}
}

func (pool *BreakerPool) GetBreaker(key string) *gobreaker.CircuitBreaker[any] {
	pool.mu.RLock()
	breaker, exist := pool.mp[key]
	pool.mu.RUnlock()
	if exist {
		return breaker
	}

	pool.mu.Lock()
	defer pool.mu.Unlock()
	if breaker, exist := pool.mp[key]; exist {
		return breaker
	}

	setting := gobreaker.Settings{
		Name:         key,
		MaxRequests:  3,
		Interval:     30 * time.Second,
		BucketPeriod: 3 * time.Second,
		Timeout:      15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			validRequests := counts.Requests
			if counts.TotalExclusions < validRequests {
				validRequests -= counts.TotalExclusions
			} else {
				validRequests = 0
			}
			if validRequests < 20 {
				return false
			}
			failRate := float64(counts.TotalFailures) / float64(validRequests)
			return failRate >= 0.5 || counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			slog.Info("Breaker State Changed", "name", name, "from", from.String(), "to", to.String())
		},
		IsSuccessful: func(err error) bool {
			if err == nil {
				return true
			}
			st, ok := status.FromError(err)
			if !ok {
				return false
			}
			switch st.Code() {
			case codes.InvalidArgument,
				codes.NotFound,
				codes.AlreadyExists,
				codes.FailedPrecondition,
				codes.PermissionDenied,
				codes.Unauthenticated:
				return true
			case codes.DeadlineExceeded,
				codes.Unavailable,
				codes.Internal,
				codes.ResourceExhausted,
				codes.Unknown:
				return false
			default:
				return false
			}
		},
		IsExcluded: func(err error) bool {
			if errors.Is(err, context.Canceled) {
				return true
			}
			st, ok := status.FromError(err)
			return ok && st.Code() == codes.Canceled
		},
	}

	breaker = gobreaker.NewCircuitBreaker[any](setting)
	pool.mp[key] = breaker
	return breaker
}

func CircuitBreakerDialOption() _grpc.DialOption {
	return _grpc.WithChainUnaryInterceptor(UnaryClientCircuitBreaker(defaultBreakerPool))
}

func UnaryClientCircuitBreaker(pool *BreakerPool) _grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *_grpc.ClientConn, invoker _grpc.UnaryInvoker, opts ..._grpc.CallOption) error {
		key := cc.Target() + ":" + method
		breaker := pool.GetBreaker(key)
		_, err := breaker.Execute(func() (any, error) {
			return nil, invoker(ctx, method, req, reply, cc, opts...)
		})

		switch {
		case errors.Is(err, gobreaker.ErrOpenState):
			return errs.ErrUnavailable
		case errors.Is(err, gobreaker.ErrTooManyRequests):
			return errs.ErrUnavailable
		default:
			return err
		}
	}
}
