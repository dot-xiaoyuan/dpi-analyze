package router

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/controllers"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/midderware"
	"github.com/gin-gonic/gin"
)

// define route

func Register(r *gin.Engine) {
	// 注册路由
	api := r.Group("/api")
	{
		// 登录接口（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Login()) // 登录接口
		}

		// 需要认证的 API 路由组
		api.Use(midderware.AuthMiddleware()) // 启用认证中间件
		{
			api.GET("/me", controllers.GetCurrentUser())
			// Dashboard
			api.GET("/dashboard", controllers.Dashboard())

			// Terminal
			terminal := api.Group("/terminal")
			{
				terminal.GET("/identification", controllers.Identification())
				terminal.Any("/useragent", controllers.UseragentRecord())
				terminal.Any("/application", controllers.Application())
			}
			// IP 操作
			api.GET("/ip/list", controllers.IPList())
			api.GET("/ip/detail", controllers.IPDetail())

			// Stream Log
			api.Match([]string{"GET", "POST"}, "/stream/list", controllers.StreamList())

			// Observer
			observer := api.Group("/observer")
			{
				observer.GET("/ttl", controllers.ObserverTTL())
				observer.GET("/mac", controllers.ObserverMac())
				observer.GET("/ua", controllers.ObserverUa())
				observer.GET("/device", controllers.ObserverDevice())
			}

			// Users
			users := api.Group("/users")
			{
				users.GET("/list", controllers.UserList())
				users.GET("/events/log", controllers.UserEventsLog())
			}
		}
	}

	// 默认路由（处理 404）
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "route not found"})
	})

}
