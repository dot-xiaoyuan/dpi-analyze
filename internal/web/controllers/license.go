package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	mongodb "github.com/dot-xiaoyuan/dpi-analyze/pkg/component/db/mongo"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/component/license"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"net/http"
	"os"
)

type LicenseRes struct {
	LicenseCode string `json:"license_code"`
	Version     string `json:"version"`
	ExpireDate  string `json:"expire_date"`
	Qrcode      string `json:"qrcode"`
	MachineID   string `json:"machine_id"`
}

func LicenseUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		type LicenseCode struct {
			Code string `json:"license_code"`
		}
		code := LicenseCode{}
		err := c.ShouldBindJSON(&code)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		config.Cfg.License = config.License{
			Sn: code.Code,
		}
		err = license.CheckLicense()
		if err != nil {
			common.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		err = mongodb.Store2Mongo()
		if err != nil {
			common.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		common.SuccessResponse(c, getLicense(c))
	}
}

func License() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := generateQrcode(); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		common.SuccessResponse(c, getLicense(c))
	}
}

func getLicense(c *gin.Context) LicenseRes {
	schema := "http"
	if c.Request.TLS != nil {
		schema = "https"
	}
	return LicenseRes{
		LicenseCode: config.Cfg.License.Sn,
		Version:     config.Version,
		ExpireDate:  config.Cfg.License.ExpireTime.Format("2006-01-02 15:04:05"),
		Qrcode:      fmt.Sprintf("%s://%s/static/license_qrcode.png", schema, c.Request.Host),
		MachineID:   config.Cfg.MachineID,
	}
}

// 生成二维码
func generateQrcode() error {
	path := fmt.Sprintf("%s/license_qrcode.png", config.UploadDir)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}
	type temp struct {
		Type string `json:"type"`
		Code string `json:"code"`
	}
	t := temp{
		Type: "srun-dpi",
		Code: config.Cfg.MachineID,
	}
	jsonData, err := json.Marshal(t)
	if err != nil {
		return err
	}
	qr, err := qrcode.New(string(jsonData), qrcode.Medium)
	if err != nil {
		return err
	}
	err = qr.WriteFile(256, path)
	if err != nil {
		return err
	}
	return nil
}
