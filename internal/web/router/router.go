package router

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/controllers"
	"github.com/gin-gonic/gin"
)

// define route

func Register(r *gin.Engine) {
	v1 := r.Group("/v1")
	v1.GET("/dashboard", controllers.Dashboard())
	v1.GET("/ip/list", controllers.IPList())
	v1.GET("/ip/detail", controllers.IPDetail())
	// stream log
	v1.Match([]string{"GET", "POST"}, "/stream/list", controllers.StreamList())
	// observer
	v1.GET("/observer/ttl", controllers.ObserverTTL())
	v1.GET("/observer/mac", controllers.ObserverMac())
	v1.GET("/observer/ua", controllers.ObserverUa())
	// users
	v1.GET("/users/list", controllers.UserList())
	v1.GET("/users/events/log", controllers.UserEventsLog())
}
