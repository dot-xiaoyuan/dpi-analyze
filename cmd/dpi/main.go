package main

import (
	"context"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/analyze"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/config"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/logger"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/capture"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetOutput(os.Stdout)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC : %v", err)
			fmt.Println("发生严重错误，请联系支持人员。")
			os.Exit(1)
		}
	}()

	// 加载配置
	err := config.LoadConfig()
	if err != nil {
		zap.L().Panic("Load config err", zap.Error(err))
	}

	// 加载日志
	logger.InitLogger(config.Cfg.LogLevel)

	// Make Context
	ctx, cancel := context.WithCancel(context.Background())

	// Packet Capture
	assembly := analyze.NewAnalyzer()
	done := make(chan struct{})
	go capture.StartCapture(ctx, capture.Config{
		OffLine: config.Cfg.Capture.OfflineFile,
		Nic:     config.Cfg.Capture.NIC,
		SnapLen: config.Cfg.Capture.SnapLen,
	}, assembly, done)

	// 信号阻塞
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-done:
		cancel()
		closed := assembly.Assembler.FlushAll()
		assembly.Factory.WaitGoRoutines()
		log.Printf("Flushed %d streams\n", closed)
	case <-signalChan:
		cancel()
		log.Println("Received an interrupt, stopping analysis...")
		time.Sleep(time.Second)
		os.Exit(0)
	}

}
