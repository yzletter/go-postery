package crontab

import (
	"log/slog"

	"github.com/robfig/cron/v3"
)

type CrontabBuilder struct {
	crontab *cron.Cron
	funcs   []func()
	specs   []string
}

func NewCrontabBuilder() *CrontabBuilder {
	return &CrontabBuilder{
		crontab: cron.New(),
		funcs:   make([]func(), 0),
		specs:   make([]string, 0),
	}
}

func (builder *CrontabBuilder) AddFuncWithSpec(spec string, f func()) *CrontabBuilder {
	builder.funcs = append(builder.funcs, f)
	builder.specs = append(builder.specs, spec)
	return builder
}

func (builder *CrontabBuilder) Build() {
	length := len(builder.funcs)

	for i := 0; i < length; i++ {
		if _, err := builder.crontab.AddFunc(builder.specs[i], builder.funcs[i]); err != nil {
			slog.Error("Crontab Add Func Failed", "error", err)
		}
	}

	builder.crontab.Start()
}

func (builder *CrontabBuilder) Stop() {
	if builder == nil || builder.crontab == nil {
		return
	}
	builder.crontab.Stop()
}
