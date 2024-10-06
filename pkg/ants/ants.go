package ants

import (
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"os"
	"sync"
)

var (
	Pool *ants.Pool
	one  sync.Once
)

func Setup(maxGoroutines int) {
	one.Do(func() {
		var err error
		Pool, err = ants.NewPool(maxGoroutines)
		if err != nil {
			zap.L().Error(err.Error())
			os.Exit(1)
		}
	})
}

func Submit(task func()) error {
	return Pool.Submit(task)
}

func Release() {
	Pool.Release()
}
