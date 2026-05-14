package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// Claims JWT Claims
type Claims struct {
	UserID       uint   `json:"user_id"`
	TenantID     uint   `json:"tenant_id"`
	Username     string `json:"username"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	jwt.RegisteredClaims
}

var jwtSecret []byte
var jwtExpireHour int
var jwtIssuer string

// InitJWT 初始化JWT配置
func InitJWT(secret string, expireHour int, issuer string) {
	jwtSecret = []byte(secret)
	jwtExpireHour = expireHour
	jwtIssuer = issuer
}

// GenerateToken 生成JWT Token
func GenerateToken(userID, tenantID uint, username string, isSuperAdmin bool) (string, error) {
	claims := Claims{
		UserID:       userID,
		TenantID:     tenantID,
		Username:     username,
		IsSuperAdmin: isSuperAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtExpireHour) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    jwtIssuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization format")
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			return
		}

		// 将用户信息注入context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("is_super_admin", claims.IsSuperAdmin)
		c.Next()
	}
}

// GetCurrentUserID 从context获取当前用户ID
func GetCurrentUserID(c *gin.Context) uint {
	userID, _ := c.Get("user_id")
	if id, ok := userID.(uint); ok {
		return id
	}
	return 0
}

// GetCurrentTenantID 从context获取当前租户ID
func GetCurrentTenantID(c *gin.Context) uint {
	tenantID, _ := c.Get("tenant_id")
	if id, ok := tenantID.(uint); ok {
		return id
	}
	return 0
}

// GetCurrentUsername 从context获取当前用户名
func GetCurrentUsername(c *gin.Context) string {
	username, _ := c.Get("username")
	if name, ok := username.(string); ok {
		return name
	}
	return ""
}

// IsSuperAdmin 判断当前用户是否为超级管理员
func IsSuperAdmin(c *gin.Context) bool {
	isAdmin, _ := c.Get("is_super_admin")
	if admin, ok := isAdmin.(bool); ok {
		return admin
	}
	return false
}
