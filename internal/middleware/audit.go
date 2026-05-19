package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/logger"
)

// AuditLog 操作审计中间件
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅记录写操作
		method := c.Request.Method
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			c.Next()
			return
		}

		// 读取body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		start := time.Now()
		c.Next()
		_ = time.Since(start)

		// 在goroutine之前提取context数据（Gin会回收context）
		userID := GetCurrentUserID(c)
		username := GetCurrentUsername(c)
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		path := c.Request.URL.Path

		// 异步记录审计日志
		go func() {
			auditLog := model.AuditLog{
				UserID:    userID,
				Username:  username,
				Action:    method + " " + path,
				Resource:  path,
				Detail:    string(bodyBytes),
				IP:        clientIP,
				UserAgent: userAgent,
			}

			if err := model.GetDB().Create(&auditLog).Error; err != nil {
				logger.Errorf("save audit log failed: %v", err)
			}
		}()
	}
}
