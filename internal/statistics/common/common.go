package common

import "github.com/gin-gonic/gin"

func Error(err error, c *gin.Context) {
	c.JSON(400, gin.H{
		"message": err.Error(),
	})
}
