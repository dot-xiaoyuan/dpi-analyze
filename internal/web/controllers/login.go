package controllers

import (
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/common"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/midderware"
	"github.com/dot-xiaoyuan/dpi-analyze/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req midderware.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "请求参数缺失")
			return
		}

		if req.Username != config.Cfg.Username || req.Password != config.Cfg.Password {
			common.ErrorResponse(c, http.StatusBadRequest, "账号或密码错误")
			return
		}

		token, err := generateJWT(req.Username)
		if err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "登录失败")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token, "machine_id": config.Cfg.MachineID})
		return
	}
}

func ChangePassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req midderware.ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			common.ErrorResponse(c, http.StatusBadRequest, "请求参数缺失")
			return
		}

		// 校验当前密码
		if req.OldPassword != config.Cfg.Password {
			common.ErrorResponse(c, http.StatusBadRequest, "当前密码错误")
			return
		}

		// 更新密码
		config.Cfg.Password = req.NewPassword
		common.SuccessResponse(c, "密码修改成功")
	}
}

func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.MustGet("username").(string)
		c.JSON(http.StatusOK, gin.H{"username": username, "machine_id": config.Cfg.MachineID})
		return
	}
}

func generateJWT(username string) (string, error) {
	claims := midderware.Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString([]byte(midderware.SecretKey))
	return s, err
}
