package controllers

import (
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

func Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
			return
		}

		filename := filepath.Base(file.Filename)
		dst := fmt.Sprintf("%s/%s", config.UploadDir, filename)
		if err = c.SaveUploadedFile(file, dst); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
		common.SuccessResponse(c, dst)
	}
}
