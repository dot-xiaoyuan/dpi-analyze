package router

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/controllers"
	"github.com/gin-gonic/gin"
)

// define route

func Register(r *gin.Engine) {
	v1 := r.Group("/v1")
	v1.GET("/dashboard", controllers.Dashboard())
}
