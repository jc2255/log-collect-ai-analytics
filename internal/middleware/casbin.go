package middleware

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

var enforcer *casbin.Enforcer

// InitCasbin 初始化Casbin
func InitCasbin(db *gorm.DB) error {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return err
	}

	e, err := casbin.NewEnforcer("configs/rbac_model.conf", adapter)
	if err != nil {
		return err
	}

	if err = e.LoadPolicy(); err != nil {
		return err
	}

	enforcer = e
	return nil
}

// GetEnforcer 获取Casbin enforcer
func GetEnforcer() *casbin.Enforcer {
	return enforcer
}

// CasbinRBAC RBAC权限中间件
func CasbinRBAC() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户
		userID := GetCurrentUserID(c)
		if userID == 0 {
			response.Unauthorized(c, "user not found")
			return
		}

		// admin用户跳过权限检查
		if GetCurrentUsername(c) == "admin" {
			c.Next()
			return
		}

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 使用用户ID作为subject进行权限检查
		sub := formatUserSub(userID)
		ok, err := enforcer.Enforce(sub, path, method)
		if err != nil {
			response.Error(c, response.ErrorCode, "permission check failed")
			c.Abort()
			return
		}

		if !ok {
			response.Forbidden(c, "no permission to access this resource")
			return
		}

		c.Next()
	}
}

func formatUserSub(userID uint) string {
	return fmt.Sprintf("user:%d", userID)
}

// AddRolePolicy 为角色添加策略
func AddRolePolicy(roleCode, path, method string) (bool, error) {
	return enforcer.AddPolicy(roleCode, path, method)
}

// RemoveRolePolicy 移除角色策略
func RemoveRolePolicy(roleCode, path, method string) (bool, error) {
	return enforcer.RemovePolicy(roleCode, path, method)
}

// AddUserRole 为用户分配角色
func AddUserRole(userID uint, roleCode string) (bool, error) {
	sub := formatUserSub(userID)
	return enforcer.AddGroupingPolicy(sub, roleCode)
}

// RemoveUserRole 移除用户角色
func RemoveUserRole(userID uint, roleCode string) (bool, error) {
	sub := formatUserSub(userID)
	return enforcer.RemoveGroupingPolicy(sub, roleCode)
}
