package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dot-xiaoyuan/dpi-analyze/internal/web/midderware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
)

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req midderware.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
			return
		}

		h := sha256.New()
		h.Write([]byte("123123"))
		sha := hex.EncodeToString(h.Sum(nil))
		fmt.Println(sha)
		if req.Username != "yuantong" || req.Password != sha {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials"})
			return
		}

		token, err := generateJWT(req.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
	}
}

func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.MustGet("username").(string)
		c.JSON(http.StatusOK, gin.H{"username": username})
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
