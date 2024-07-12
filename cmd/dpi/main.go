package main

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/config"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/logger"
	"log"
	"os"
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
		panic(err)
	}

	// 加载日志
	logger.InitLogger(config.Cfg.LogLevel)
}
