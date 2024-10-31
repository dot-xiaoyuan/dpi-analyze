package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/router"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// web server

//go:embed dist
var dist embed.FS

var (
	web *gin.Engine
)

type Config struct {
	Port uint
}

func NewWebServer(c Config) {
	zap.L().Info(i18n.T("Start Load Mongodb Component"))
	sp := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	sp.Start()
	if err := mongo.Setup(); err != nil {
		sp.Stop()
		os.Exit(1)
	}
	sp.Stop()
	zap.L().Info(i18n.T("Starting Web Server"))
	web = gin.Default()
	web.Use(ServerStatic("dist", dist))
	// cors
	web.Use(cors.Default())
	web.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许React前端所在的域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// 注册路由
	router.Register(web)
	// 日志中间件
	//web.Use(logger.GinLogger())
	// 服务
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Port),
		Handler: web,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Fatal(i18n.T("Failed to start Web Server"), zap.Error(err))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
	zap.L().Info(i18n.T("Shutting down Web Server"))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		zap.L().Fatal(i18n.T("Failed to shutdown Web Server"), zap.Error(err))
	}
	zap.L().Info(i18n.T("Shut down Web Server"))
}

func ServerStatic(prefix string, embedFs embed.FS) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 去掉前缀
		fsys, err := fs.Sub(embedFs, prefix)
		if err != nil {
			panic(err)
		}
		fs2 := http.FS(fsys)
		f, err := fs2.Open(ctx.Request.URL.Path)
		if err != nil {
			// 判断文件不存在，退出交给其他路由函数
			ctx.Next()
			return
		}
		defer f.Close()
		http.FileServer(fs2).ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Abort()
	}
}
