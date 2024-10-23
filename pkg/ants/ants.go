package ants

import (
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"runtime"
	"sync"
	"time"
)

var (
	Pool *ants.Pool
	one  sync.Once
)

func Setup(maxGoroutines int) error {
	var err error
	one.Do(func() {
		poolSize := runtime.NumCPU() * 2 // 线程池大小建议设置为 CPU 核心数的2倍
		zap.L().Info("按照CPU核心数设置线程池大小", zap.Int("poolSize", poolSize))
		Pool, err = ants.NewPool(maxGoroutines, ants.WithExpiryDuration(10*time.Second))
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
