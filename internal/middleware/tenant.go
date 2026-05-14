package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// TenantIsolation 多租户数据隔离中间件
func TenantIsolation() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := GetCurrentTenantID(c)
		if tenantID == 0 && !IsSuperAdmin(c) {
			response.Error(c, response.ErrorCode, "tenant not found")
			c.Abort()
			return
		}
		c.Next()
	}
}

// TenantScope 获取租户作用域的数据库查询(自动添加tenant_id条件)
func TenantScope(c *gin.Context) func(*model.BaseModel) {
	tenantID := GetCurrentTenantID(c)
	if IsSuperAdmin(c) && tenantID == 0 {
		return nil // 超级管理员不限制租户
	}
	return nil
}

// GetTenantDB 获取带租户隔离的数据库实例
func GetTenantDB(c *gin.Context) *model.TenantDB {
	tenantID := GetCurrentTenantID(c)
	isSuperAdmin := IsSuperAdmin(c)
	return &model.TenantDB{
		TenantID:     tenantID,
		IsSuperAdmin: isSuperAdmin,
	}
}
