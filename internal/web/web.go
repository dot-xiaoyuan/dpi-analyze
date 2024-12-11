package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/router"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/i18n"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// web server

//go:embed build
var build embed.FS

var (
	web *gin.Engine
)

type Config struct {
	Port uint
}

func NewWebServer(c Config) {
	zap.L().Info(i18n.T("Start Load Mongodb Component"))
	if err := mongo.Setup(); err != nil {
		os.Exit(1)
	}
	zap.L().Info(i18n.T("Starting Web Server"))
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	web = gin.Default()
	// cors
	web.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 允许React前端所在的域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// 静态资源
	web.Static("/static", config.UploadDir)
	// 注册路由
	router.Register(web)
	web.Use(ServerStatic("build", build))
	// 日志中间件
	//web.Use(logger.GinLogger())
	// 服务
	addr := fmt.Sprintf(":%d", c.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: web,
	}

	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}
	erred := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
	log.Println("server start success", erred)
	if erred != nil {
		panic(err)
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
		fsys, err := fs.Sub(embedFs, prefix)
		if err != nil {
			panic(err)
		}
		fs2 := http.FS(fsys)
		requestedPath := ctx.Request.URL.Path

		// 打开请求的路径文件
		f, err := fs2.Open(requestedPath)
		if err != nil {
			// 如果文件不存在，返回 index.html
			if errors.Is(err, fs.ErrNotExist) {
				indexFile, err := fs2.Open("index.html")
				if err != nil {
					ctx.String(http.StatusNotFound, "404 page not found")
					return
				}
				defer indexFile.Close()
				http.ServeContent(ctx.Writer, ctx.Request, "index.html", time.Now(), indexFile)
				ctx.Abort()
				return
			}
			// 其他错误
			ctx.Next()
			return
		}

		// 文件存在，直接提供服务
		defer f.Close()
		http.FileServer(fs2).ServeHTTP(ctx.Writer, ctx.Request)
		ctx.Abort()
	}
}
