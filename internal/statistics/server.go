package statistics

import (
	"context"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/statistics/handlers"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Config struct {
	Port uint
}

func StartStatistics(c Config) {
	zap.L().Info(i18n.T("Starting Statistics"))
	r := gin.Default()
	setupRoutes(r)

	r.Use(logger.GinLogger())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: r,
	}

	// run http server
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal(i18n.T("Failed to start server"), zap.Error(err))
		}
	}()

	// Wait for context cancellation
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	zap.L().Info(i18n.T("Statistics Stopped"))

	// Make Context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zap.L().Info(i18n.T("Server Shutdown"), zap.Error(err))
	}
	zap.L().Info(i18n.T("Server exiting"))
}

func setupRoutes(r *gin.Engine) {
	v1 := r.Group("/v1")
	v1.GET("/index", Index)
	v1.GET("/ttl", handlers.TTL())
	v1.GET("/mac", handlers.Mac())
	v1.GET("/ip-tables", handlers.IpTables())
	v1.GET("/dashboard", handlers.Dashboard())
}
