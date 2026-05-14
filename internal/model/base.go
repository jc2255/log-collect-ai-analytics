package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Tenant 租户
type Tenant struct {
	BaseModel
	Name        string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Code        string `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Status      int8   `gorm:"default:1;comment:1-启用 0-禁用" json:"status"`
	QuotaConfig string `gorm:"type:text;comment:配额配置JSON" json:"quota_config"`
	Description string `gorm:"size:500" json:"description"`
}

func (Tenant) TableName() string { return "tenants" }

// User 用户
type User struct {
	BaseModel
	TenantID     uint   `gorm:"not null;index" json:"tenant_id"`
	Username     string `gorm:"size:50;not null" json:"username"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	Nickname     string `gorm:"size:100" json:"nickname"`
	Email        string `gorm:"size:100" json:"email"`
	Phone        string `gorm:"size:20" json:"phone"`
	Avatar       string `gorm:"size:500" json:"avatar"`
	Status       int8   `gorm:"default:1;comment:1-启用 0-禁用" json:"status"`
	IsSuperAdmin bool   `gorm:"default:false" json:"is_super_admin"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Roles  []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

// Role 角色
type Role struct {
	BaseModel
	TenantID    uint   `gorm:"not null;index" json:"tenant_id"`
	Name        string `gorm:"size:50;not null" json:"name"`
	Code        string `gorm:"size:50;not null" json:"code"`
	Description string `gorm:"size:500" json:"description"`
	Status      int8   `gorm:"default:1" json:"status"`

	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

func (Role) TableName() string { return "roles" }

// Permission 权限
type Permission struct {
	BaseModel
	Name     string `gorm:"size:100;not null" json:"name"`
	Code     string `gorm:"size:100;not null;uniqueIndex" json:"code"`
	Type     string `gorm:"size:20;not null;comment:menu/button/api" json:"type"`
	ParentID uint   `gorm:"default:0" json:"parent_id"`
	Path     string `gorm:"size:200;comment:API路径或前端路由" json:"path"`
	Method   string `gorm:"size:10;comment:HTTP方法" json:"method"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Icon     string `gorm:"size:100" json:"icon"`
	Status   int8   `gorm:"default:1" json:"status"`
}

func (Permission) TableName() string { return "permissions" }

// Menu 菜单
type Menu struct {
	BaseModel
	ParentID  uint   `gorm:"default:0" json:"parent_id"`
	Name      string `gorm:"size:50;not null" json:"name"`
	Path      string `gorm:"size:200" json:"path"`
	Component string `gorm:"size:200" json:"component"`
	Icon      string `gorm:"size:100" json:"icon"`
	Sort      int    `gorm:"default:0" json:"sort"`
	Hidden    bool   `gorm:"default:false" json:"hidden"`
	Status    int8   `gorm:"default:1" json:"status"`
	PermCode  string `gorm:"size:100;comment:关联权限code" json:"perm_code"`
}

func (Menu) TableName() string { return "menus" }
