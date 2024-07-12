package main

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/logger"
)

func main() {
	logger.InitLogger("debug")
	logger.Debug("debug")
}
