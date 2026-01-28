package graceful_stop

import (
	"log/slog"
	"os"
	"os/signal"
)

type GracefulStopBuilder struct {
	signals []os.Signal // 监听的信号
	funcs   []func()    // 退出前要执行的函数
}

func NewGracefulStopBuilder() *GracefulStopBuilder {
	signals := make([]os.Signal, 0)
	funcs := make([]func(), 0)
	return &GracefulStopBuilder{
		signals: signals,
		funcs:   funcs,
	}
}

func (builder *GracefulStopBuilder) NotifySignal(signal os.Signal) *GracefulStopBuilder {
	builder.signals = append(builder.signals, signal)
	return builder
}

func (builder *GracefulStopBuilder) AddFunc(f func()) *GracefulStopBuilder {
	builder.funcs = append(builder.funcs, f)
	return builder
}

func (builder *GracefulStopBuilder) Build() {
	var listen func()

	// 监听函数
	listen = func() {
		if len(builder.signals) == 0 {
			return
		}

		ch := make(chan os.Signal, 1)

		// 注册信号
		signal.Notify(ch, builder.signals...)

		// 阻塞, 直到信号到来
		s := <-ch

		slog.Info("信号 " + s.String() + " 成功监听, 开始优雅退出")

		// 退出前具体要做的工作
		for _, _func := range builder.funcs {
			_func()
		}

		slog.Info("退出前所有任务完成")

		// 退出所有进程
		os.Exit(0)
	}

	// 开始监听信号
	go listen()
}
