package common

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
	var res any
	_ = json.Unmarshal(data.([]byte), &res)

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    res,
	})
}

func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}
