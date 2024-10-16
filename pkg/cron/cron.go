package cron

import (
	"github.com/robfig/cron/v3"
	"sync"
)

var (
	one  sync.Once
	Cron *cron.Cron
)

func Setup() error {
	one.Do(func() {
		loadCronClient()
	})
	return nil
}

func loadCronClient() {
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
