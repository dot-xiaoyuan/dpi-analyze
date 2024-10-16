package cron

import (
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"os"
	"sync"
)

var (
	one  sync.Once
	Cron *cron.Cron
)

func Setup() {
	one.Do(func() {
		loadCronClient()
	})
}

func loadCronClient() {
	defer func() {
		if err := recover(); err != nil {
			zap.L().Error(i18n.T("Failed to load cron"), zap.Any("error", err))
			os.Exit(1)
		}
	}()

	Cron = cron.New()
}

func AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	return Cron.AddJob(spec, cmd)
}

func Start() {
	Cron.Start()
}

func Stop() {
	Cron.Stop()
}
