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
			api.POST("/change-password", controllers.ChangePassword())

			api.GET("/me", controllers.GetCurrentUser())
			// Dashboard
			api.GET("/dashboard", controllers.Dashboard())
			// 上传
			api.POST("/upload", controllers.Upload())

			// Terminal 终端相关
			terminal := api.Group("/terminal")
			{
				terminal.POST("/identification", controllers.Identification())
				terminal.POST("/useragent", controllers.UseragentRecord())
				terminal.POST("/application", controllers.Application())
				terminal.POST("/detail", controllers.Detail())
			}
			// Judge 特征判定
			feature := api.Group("/feature")
			{
				judge := feature.Group("/judge")
				{
					judge.POST("/realtime", controllers.FeatureJudgement())
					judge.POST("/suspected", controllers.JudgeSuspected())
				}
				feature.POST("/library", controllers.FeatureLibrary())
				feature.PUT("/library", controllers.FeatureUpdate())
			}

			// policy 策略配置
			policy := api.Group("/policy")
			{
				policy.GET("/list", controllers.PolicyList())
				policy.POST("/update", controllers.PolicyUpdate())
			}

			// log 日志管理
			log := api.Group("/log")
			{
				log.POST("/users/events", controllers.UserEventsLog())
				log.POST("/proxy", controllers.JudgeRealtime())
			}

			// setting 系统配置
			settings := api.Group("/setting")
			{
				settings.GET("/config", controllers.ConfigList())
				settings.PUT("/config", controllers.ConfigUpdate())
			}

			// IP 操作
			//api.GET("/ip/list", controllers.IPList())
			//api.GET("/ip/detail", controllers.IPDetail())

			// Observer
			//observer := api.Group("/observer")
			//{
			//	observer.GET("/ttl", controllers.ObserverTTL())
			//	observer.GET("/mac", controllers.ObserverMac())
			//	observer.GET("/ua", controllers.ObserverUa())
			//	observer.GET("/device", controllers.ObserverDevice())
			//}

			// Users
			users := api.Group("/users")
			{
				users.GET("/list", controllers.UserList())
				users.GET("/events/log", controllers.UserEventsLog())
			}

			// Licence 授权
			api.GET("/license", controllers.License())
			api.PUT("/license", controllers.LicenseUpdate())
		}
	}

	// 默认路由（处理 404）
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "route not found"})
	})

}
