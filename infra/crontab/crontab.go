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
	crontab := cron.New()
	funcs := make([]func(), 0)
	specs := make([]string, 0)
	return &CrontabBuilder{
		crontab: crontab,
		funcs:   funcs,
		specs:   specs,
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
		_, err := builder.crontab.AddFunc(builder.specs[i], builder.funcs[i])
		if err != nil {
			slog.Error("crontab add func failed", "error", err)
		}
	}

	// 启动 Crontab
	builder.crontab.Start()
}
