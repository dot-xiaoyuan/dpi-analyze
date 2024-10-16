package ants

import (
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"sync"
)

var (
	Pool *ants.Pool
	one  sync.Once
)

func Setup(maxGoroutines int) error {
	var err error
	one.Do(func() {
		Pool, err = ants.NewPool(maxGoroutines)
		if err != nil {
			zap.L().Error(err.Error())
		}
	})
	return err
}

func Submit(task func()) error {
	return Pool.Submit(task)
}

func Release() {
	Pool.Release()
}
