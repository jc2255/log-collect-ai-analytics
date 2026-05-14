package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
	"github.com/cj/log-collect-ai-analytics/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.Login(&req)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetUserInfo 获取当前用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	info, err := h.authService.GetUserInfo(userID)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, info)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetCurrentUserID(c)
	if err := h.authService.ChangePassword(userID, &req); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

// UserHandler 用户管理处理器
type UserHandler struct {
	authService *service.AuthService
	rbacService *service.RBACService
}

func NewUserHandler(authService *service.AuthService, rbacService *service.RBACService) *UserHandler {
	return &UserHandler{authService: authService, rbacService: rbacService}
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req service.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误: "+err.Error())
		return
	}

	// 非超级管理员只能创建本租户用户
	if !middleware.IsSuperAdmin(c) {
		req.TenantID = middleware.GetCurrentTenantID(c)
	}

	user, err := h.authService.CreateUser(&req)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, user)
}

// ListUsers 用户列表
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	tenantID := middleware.GetCurrentTenantID(c)
	if middleware.IsSuperAdmin(c) {
		if tid := c.Query("tenant_id"); tid != "" {
			id, _ := strconv.Atoi(tid)
			tenantID = uint(id)
		}
	}

	users, total, err := h.authService.ListUsers(tenantID, page, pageSize)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.PageSuccess(c, users, total, page, pageSize)
}

// AssignRoles 为用户分配角色
func (h *UserHandler) AssignRoles(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误")
		return
	}

	if err := h.rbacService.AssignUserRoles(uint(id), req.RoleIDs); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}
