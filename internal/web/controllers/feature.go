package controllers

import (
	"encoding/json"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/socket/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FeatureLibrary() gin.HandlerFunc {
	return func(c *gin.Context) {
		bytes, err := socket.SendUnixMessage(socket.FeatureLibrary, nil)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res any
		_ = json.Unmarshal(bytes, &res)
		common.SuccessResponse(c, res)
		return
	}
}

type FeatureUpdateRequest struct {
	Filepath string `json:"filepath"`
	Module   string `json:"module"`
}

func FeatureUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req FeatureUpdateRequest
		if err := c.BindJSON(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		bytes, err := socket.SendUnixMessage(socket.FeatureUpdate, req)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		var res models.Response
		_ = json.Unmarshal(bytes, &res)
		c.JSON(http.StatusOK, res)
		return
	}
}
