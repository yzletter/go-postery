package graceful_stop

import (
	"os"
	"os/signal"
)

// GracefulStopBuilder 优雅退出 —— 构造器模式
type GracefulStopBuilder struct {
	signals   []os.Signal // 监听的信号
	functions []func()    // 退出前要执行的函数
}

func NewGracefulStopBuilder() *GracefulStopBuilder {
	signals := make([]os.Signal, 0)
	functions := make([]func(), 0)
	return &GracefulStopBuilder{
		signals:   signals,
		functions: functions,
	}
}

// NotifySignal 注册信号
func (builder *GracefulStopBuilder) NotifySignal(signal os.Signal) *GracefulStopBuilder {
	builder.signals = append(builder.signals, signal)
	return builder
}

// AddFunc 添加函数
func (builder *GracefulStopBuilder) AddFunc(f func()) *GracefulStopBuilder {
	builder.functions = append(builder.functions, f)
	return builder
}

// Build 非阻塞型
func (builder *GracefulStopBuilder) Build() {
	// 开始监听信号
	go builder.build()()
}

// BuildBlock 阻塞型
func (builder *GracefulStopBuilder) BuildBlock() {
	// 开始监听信号
	builder.build()()
}

func (builder *GracefulStopBuilder) build() func() {
	listen := func() {
		if len(builder.signals) == 0 {
			return
		}

		ch := make(chan os.Signal, 1)

		// 监听全部信号
		signal.Notify(ch, builder.signals...)

		<-ch

		// 执行退出前所有任务
		for _, _func := range builder.functions {
			_func()
		}

		// 退出进程
		os.Exit(0)
	}
	return listen
}
