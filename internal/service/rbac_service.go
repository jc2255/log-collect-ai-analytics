package service

import (
	"fmt"

	"github.com/cj/log-collect-ai-analytics/internal/dao"
	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
)

// RBACService RBAC服务
type RBACService struct {
	roleDAO *dao.RoleDAO
	permDAO *dao.PermissionDAO
	menuDAO *dao.MenuDAO
	userDAO *dao.UserDAO
}

func NewRBACService(roleDAO *dao.RoleDAO, permDAO *dao.PermissionDAO, menuDAO *dao.MenuDAO, userDAO *dao.UserDAO) *RBACService {
	return &RBACService{
		roleDAO: roleDAO,
		permDAO: permDAO,
		menuDAO: menuDAO,
		userDAO: userDAO,
	}
}

// --- 角色管理 ---

type CreateRoleRequest struct {
	TenantID    uint   `json:"tenant_id"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

func (s *RBACService) CreateRole(req *CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{
		TenantID:    req.TenantID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Status:      1,
	}
	err := s.roleDAO.Create(role)
	return role, err
}

func (s *RBACService) GetRole(id uint) (*model.Role, error) {
	return s.roleDAO.GetByID(id)
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

func (s *RBACService) UpdateRole(id uint, req *UpdateRoleRequest) error {
	role, err := s.roleDAO.GetByID(id)
	if err != nil {
		return err
	}
	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Code != "" {
		role.Code = req.Code
	}
	if req.Description != "" {
		role.Description = req.Description
	}
	if req.Status != 0 {
		role.Status = req.Status
	}
	return s.roleDAO.Update(role)
}

func (s *RBACService) DeleteRole(id uint) error {
	return s.roleDAO.Delete(id)
}

func (s *RBACService) ListRoles(tenantID uint, page, pageSize int) ([]model.Role, int64, error) {
	return s.roleDAO.List(tenantID, page, pageSize)
}

// AssignPermissions 为角色分配权限
func (s *RBACService) AssignPermissions(roleID uint, permIDs []uint) error {
	// 更新数据库关联
	if err := s.roleDAO.AssignPermissions(roleID, permIDs); err != nil {
		return err
	}

	// 更新Casbin策略
	role, err := s.roleDAO.GetByID(roleID)
	if err != nil {
		return err
	}

	// 清除旧策略
	enforcer := middleware.GetEnforcer()
	if enforcer != nil {
		enforcer.RemoveFilteredPolicy(0, role.Code)
		// 添加新策略
		for _, perm := range role.Permissions {
			if perm.Type == "api" && perm.Path != "" {
				enforcer.AddPolicy(role.Code, perm.Path, perm.Method)
			}
		}
		enforcer.SavePolicy()
	}
	return nil
}

// AssignUserRoles 为用户分配角色
func (s *RBACService) AssignUserRoles(userID uint, roleIDs []uint) error {
	if err := s.userDAO.AssignRoles(userID, roleIDs); err != nil {
		return err
	}

	// 更新Casbin用户角色关联
	enforcer := middleware.GetEnforcer()
	if enforcer != nil {
		sub := fmt.Sprintf("user:%d", userID)
		enforcer.RemoveFilteredGroupingPolicy(0, sub)
		for _, roleID := range roleIDs {
			role, err := s.roleDAO.GetByID(roleID)
			if err == nil {
				enforcer.AddGroupingPolicy(sub, role.Code)
			}
		}
		enforcer.SavePolicy()
	}
	return nil
}

// --- 权限管理 ---

type CreatePermissionRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Type     string `json:"type" binding:"required"`
	ParentID uint   `json:"parent_id"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	Sort     int    `json:"sort"`
	Icon     string `json:"icon"`
}

func (s *RBACService) CreatePermission(req *CreatePermissionRequest) (*model.Permission, error) {
	perm := &model.Permission{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		ParentID: req.ParentID,
		Path:     req.Path,
		Method:   req.Method,
		Sort:     req.Sort,
		Icon:     req.Icon,
		Status:   1,
	}
	err := s.permDAO.Create(perm)
	return perm, err
}

func (s *RBACService) ListPermissions() ([]model.Permission, error) {
	return s.permDAO.ListAll()
}

func (s *RBACService) DeletePermission(id uint) error {
	return s.permDAO.Delete(id)
}

// --- 菜单管理 ---

type CreateMenuRequest struct {
	ParentID  uint   `json:"parent_id"`
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path"`
	Component string `json:"component"`
	Icon      string `json:"icon"`
	Sort      int    `json:"sort"`
	Hidden    bool   `json:"hidden"`
	PermCode  string `json:"perm_code"`
}

func (s *RBACService) CreateMenu(req *CreateMenuRequest) (*model.Menu, error) {
	menu := &model.Menu{
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Hidden:    req.Hidden,
		PermCode:  req.PermCode,
		Status:    1,
	}
	err := s.menuDAO.Create(menu)
	return menu, err
}

func (s *RBACService) ListMenus() ([]model.Menu, error) {
	return s.menuDAO.ListAll()
}

func (s *RBACService) UpdateMenu(id uint, req *CreateMenuRequest) error {
	menu, err := s.menuDAO.GetByID(id)
	if err != nil {
		return err
	}
	menu.ParentID = req.ParentID
	menu.Name = req.Name
	menu.Path = req.Path
	menu.Component = req.Component
	menu.Icon = req.Icon
	menu.Sort = req.Sort
	menu.Hidden = req.Hidden
	menu.PermCode = req.PermCode
	return s.menuDAO.Update(menu)
}

func (s *RBACService) DeleteMenu(id uint) error {
	return s.menuDAO.Delete(id)
}
