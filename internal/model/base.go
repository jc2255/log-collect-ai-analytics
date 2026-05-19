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

// User 用户
type User struct {
	BaseModel
	Username     string `gorm:"size:50;not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	Nickname     string `gorm:"size:100" json:"nickname"`
	Email        string `gorm:"size:100" json:"email"`
	Phone        string `gorm:"size:20" json:"phone"`
	Avatar       string `gorm:"size:500" json:"avatar"`
	Status       int8   `gorm:"default:1;comment:1-启用 0-禁用" json:"status"`
	DeptID       uint   `gorm:"default:0" json:"dept_id"`
	PostID       uint   `gorm:"default:0" json:"post_id"`
	Remark       string `gorm:"size:500" json:"remark"`

	Dept  *Department `gorm:"-" json:"dept,omitempty"`
	Post  *Post       `gorm:"-" json:"post,omitempty"`
	Roles []Role      `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

func (User) TableName() string { return "users" }

// Role 角色
type Role struct {
	BaseModel
	Name        string `gorm:"size:50;not null" json:"name"`
	Code        string `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Sort        int    `gorm:"default:0" json:"sort"`
	Status      int8   `gorm:"default:1" json:"status"`
	Description string `gorm:"size:500" json:"description"`

	Menus []Menu `gorm:"many2many:role_menus" json:"menus,omitempty"`
}

func (Role) TableName() string { return "roles" }

// Department 部门
type Department struct {
	BaseModel
	ParentID uint   `gorm:"default:0" json:"parent_id"`
	Name     string `gorm:"size:100;not null" json:"name"`
	Sort     int    `gorm:"default:0" json:"sort"`
	Leader   string `gorm:"size:50" json:"leader"`
	Phone    string `gorm:"size:20" json:"phone"`
	Email    string `gorm:"size:100" json:"email"`
	Status   int8   `gorm:"default:1" json:"status"`

	Children []Department `gorm:"-" json:"children,omitempty"`
}

func (Department) TableName() string { return "departments" }

// Post 岗位
type Post struct {
	BaseModel
	Name   string `gorm:"size:100;not null" json:"name"`
	Code   string `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Sort   int    `gorm:"default:0" json:"sort"`
	Status int8   `gorm:"default:1" json:"status"`
	Remark string `gorm:"size:500" json:"remark"`
}

func (Post) TableName() string { return "posts" }

// Menu 菜单
type Menu struct {
	BaseModel
	ParentID  uint   `gorm:"default:0" json:"parent_id"`
	Name      string `gorm:"size:50;not null" json:"name"`
	Path      string `gorm:"size:200" json:"path"`
	Component string `gorm:"size:200" json:"component"`
	Icon      string `gorm:"size:100" json:"icon"`
	Sort      int    `gorm:"default:0" json:"sort"`
	MenuType  string `gorm:"size:10;default:M;comment:M-目录 C-菜单 F-按钮" json:"menu_type"`
	Visible   int8   `gorm:"default:1;comment:1-显示 0-隐藏" json:"visible"`
	Status    int8   `gorm:"default:1" json:"status"`
	Perms     string `gorm:"size:200;comment:权限标识" json:"perms"`
	ApiPath   string `gorm:"size:200;comment:API接口路径" json:"api_path"`

	Children []Menu `gorm:"-" json:"children,omitempty"`
}

func (Menu) TableName() string { return "menus" }
