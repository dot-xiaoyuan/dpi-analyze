package router

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/controllers"
	"github.com/gin-gonic/gin"
)

// define route

func Register(r *gin.Engine) {
	api := r.Group("/api")
	api.GET("/dashboard", controllers.Dashboard())
	api.GET("/ip/list", controllers.IPList())
	api.GET("/ip/detail", controllers.IPDetail())
	// stream log
	api.Match([]string{"GET", "POST"}, "/stream/list", controllers.StreamList())
	// observer
	api.GET("/observer/ttl", controllers.ObserverTTL())
	api.GET("/observer/mac", controllers.ObserverMac())
	api.GET("/observer/ua", controllers.ObserverUa())
	api.GET("/observer/device", controllers.ObserverDevice())
	// users
	api.GET("/users/list", controllers.UserList())
	api.GET("/users/events/log", controllers.UserEventsLog())
}
