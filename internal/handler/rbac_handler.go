package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
	"github.com/cj/log-collect-ai-analytics/internal/service"
)

// RoleHandler 角色处理器
type RoleHandler struct {
	rbacService *service.RBACService
}

func NewRoleHandler(rbacService *service.RBACService) *RoleHandler {
	return &RoleHandler{rbacService: rbacService}
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误: "+err.Error())
		return
	}
	if !middleware.IsSuperAdmin(c) {
		req.TenantID = middleware.GetCurrentTenantID(c)
	}
	role, err := h.rbacService.CreateRole(&req)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, role)
}

func (h *RoleHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	role, err := h.rbacService.GetRole(uint(id))
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误")
		return
	}
	if err := h.rbacService.UpdateRole(uint(id), &req); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.rbacService.DeleteRole(uint(id)); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	tenantID := middleware.GetCurrentTenantID(c)
	roles, total, err := h.rbacService.ListRoles(tenantID, page, pageSize)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.PageSuccess(c, roles, total, page, pageSize)
}

func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误")
		return
	}
	if err := h.rbacService.AssignPermissions(uint(id), req.PermissionIDs); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantService *service.TenantService
}

func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

func (h *TenantHandler) Create(c *gin.Context) {
	var req service.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误: "+err.Error())
		return
	}
	tenant, err := h.tenantService.Create(&req)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, tenant)
}

func (h *TenantHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tenant, err := h.tenantService.GetByID(uint(id))
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, tenant)
}

func (h *TenantHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req service.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误")
		return
	}
	if err := h.tenantService.Update(uint(id), &req); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TenantHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.tenantService.Delete(uint(id)); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *TenantHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	tenants, total, err := h.tenantService.List(page, pageSize)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.PageSuccess(c, tenants, total, page, pageSize)
}

// MenuHandler 菜单处理器
type MenuHandler struct {
	rbacService *service.RBACService
}

func NewMenuHandler(rbacService *service.RBACService) *MenuHandler {
	return &MenuHandler{rbacService: rbacService}
}

func (h *MenuHandler) Create(c *gin.Context) {
	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误: "+err.Error())
		return
	}
	menu, err := h.rbacService.CreateMenu(&req)
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, menu)
}

func (h *MenuHandler) List(c *gin.Context) {
	menus, err := h.rbacService.ListMenus()
	if err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, menus)
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req service.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "参数错误")
		return
	}
	if err := h.rbacService.UpdateMenu(uint(id), &req); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.rbacService.DeleteMenu(uint(id)); err != nil {
		response.Error(c, response.ErrorCode, err.Error())
		return
	}
	response.Success(c, nil)
}
