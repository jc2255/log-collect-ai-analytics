package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
)

// 授权检查白名单路径(无需授权即可访问)
var licenseWhitelist = []string{
	"/api/v1/auth/login",
	"/api/v1/captcha",
	"/api/v1/license/",
	"/api/v1/auth/userinfo",
	"/health",
}

// licenseDB 授权检查使用的数据库实例
var licenseDB *gorm.DB

// InitLicenseDB 初始化授权检查数据库
func InitLicenseDB(db *gorm.DB) {
	licenseDB = db
}

// LicenseCheck 授权码检查中间件
// 登录后的请求需检查是否已绑定有效授权码
func LicenseCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 白名单路径放行
		for _, prefix := range licenseWhitelist {
			if strings.HasPrefix(path, prefix) || path == prefix {
				c.Next()
				return
			}
		}

		// 优先从gin context获取license状态(由handler设置)
		if licenseStatus, exists := c.Get("license_valid"); exists {
			if valid, ok := licenseStatus.(bool); ok && valid {
				c.Next()
				return
			}
		}

		// 从数据库检查是否有有效授权码
		if licenseDB != nil {
			var count int64
			licenseDB.Model(&model.License{}).Where("status = ?", 1).Count(&count)
			if count > 0 {
				c.Set("license_valid", true)
				c.Next()
				return
			}
		}

		// 默认拒绝
		c.JSON(http.StatusForbidden, gin.H{
			"code":    40301,
			"message": "license_required",
			"data":    nil,
		})
		c.Abort()
	}
}
