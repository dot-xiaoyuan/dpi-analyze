package cron

import (
	"fmt"
	v3 "github.com/robfig/cron/v3"
	"sync"
)

var Cron cron

type cron struct {
	one         sync.Once
	initialized bool
	cron        *v3.Cron
}

func (c *cron) Setup() error {
	var setupErr error
	c.one.Do(func() {
		if c.initialized {
			setupErr = fmt.Errorf("cron already initialized")
			return
		}
		c.cron = v3.New()
		c.initialized = true
	})
	return setupErr
}

func AddJob(spec string, cmd v3.Job) (v3.EntryID, error) {
	return Cron.cron.AddJob(spec, cmd)
}

func Start() {
	Cron.cron.Start()
}

func Stop() {
	Cron.cron.Stop()
}
