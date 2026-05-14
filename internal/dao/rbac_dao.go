package dao

import (
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"gorm.io/gorm"
)

// RoleDAO 角色数据访问
type RoleDAO struct {
	db *gorm.DB
}

func NewRoleDAO(db *gorm.DB) *RoleDAO {
	return &RoleDAO{db: db}
}

func (d *RoleDAO) Create(role *model.Role) error {
	return d.db.Create(role).Error
}

func (d *RoleDAO) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := d.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (d *RoleDAO) Update(role *model.Role) error {
	return d.db.Save(role).Error
}

func (d *RoleDAO) Delete(id uint) error {
	return d.db.Delete(&model.Role{}, id).Error
}

func (d *RoleDAO) List(tenantID uint, page, pageSize int) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64
	query := d.db.Model(&model.Role{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	query.Count(&total)
	err := query.Preload("Permissions").Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").Find(&roles).Error
	return roles, total, err
}

func (d *RoleDAO) AssignPermissions(roleID uint, permIDs []uint) error {
	var role model.Role
	role.ID = roleID
	var perms []model.Permission
	for _, id := range permIDs {
		perms = append(perms, model.Permission{BaseModel: model.BaseModel{ID: id}})
	}
	return d.db.Model(&role).Association("Permissions").Replace(perms)
}

// TenantDAO 租户数据访问
type TenantDAO struct {
	db *gorm.DB
}

func NewTenantDAO(db *gorm.DB) *TenantDAO {
	return &TenantDAO{db: db}
}

func (d *TenantDAO) Create(tenant *model.Tenant) error {
	return d.db.Create(tenant).Error
}

func (d *TenantDAO) GetByID(id uint) (*model.Tenant, error) {
	var tenant model.Tenant
	err := d.db.First(&tenant, id).Error
	return &tenant, err
}

func (d *TenantDAO) GetByCode(code string) (*model.Tenant, error) {
	var tenant model.Tenant
	err := d.db.Where("code = ?", code).First(&tenant).Error
	return &tenant, err
}

func (d *TenantDAO) Update(tenant *model.Tenant) error {
	return d.db.Save(tenant).Error
}

func (d *TenantDAO) Delete(id uint) error {
	return d.db.Delete(&model.Tenant{}, id).Error
}

func (d *TenantDAO) List(page, pageSize int) ([]model.Tenant, int64, error) {
	var tenants []model.Tenant
	var total int64
	d.db.Model(&model.Tenant{}).Count(&total)
	err := d.db.Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").Find(&tenants).Error
	return tenants, total, err
}

// PermissionDAO 权限数据访问
type PermissionDAO struct {
	db *gorm.DB
}

func NewPermissionDAO(db *gorm.DB) *PermissionDAO {
	return &PermissionDAO{db: db}
}

func (d *PermissionDAO) Create(perm *model.Permission) error {
	return d.db.Create(perm).Error
}

func (d *PermissionDAO) GetByID(id uint) (*model.Permission, error) {
	var perm model.Permission
	err := d.db.First(&perm, id).Error
	return &perm, err
}

func (d *PermissionDAO) Update(perm *model.Permission) error {
	return d.db.Save(perm).Error
}

func (d *PermissionDAO) Delete(id uint) error {
	return d.db.Delete(&model.Permission{}, id).Error
}

func (d *PermissionDAO) ListAll() ([]model.Permission, error) {
	var perms []model.Permission
	err := d.db.Order("sort ASC, id ASC").Find(&perms).Error
	return perms, err
}

func (d *PermissionDAO) ListByType(permType string) ([]model.Permission, error) {
	var perms []model.Permission
	err := d.db.Where("type = ?", permType).Order("sort ASC").Find(&perms).Error
	return perms, err
}

// MenuDAO 菜单数据访问
type MenuDAO struct {
	db *gorm.DB
}

func NewMenuDAO(db *gorm.DB) *MenuDAO {
	return &MenuDAO{db: db}
}

func (d *MenuDAO) Create(menu *model.Menu) error {
	return d.db.Create(menu).Error
}

func (d *MenuDAO) GetByID(id uint) (*model.Menu, error) {
	var menu model.Menu
	err := d.db.First(&menu, id).Error
	return &menu, err
}

func (d *MenuDAO) Update(menu *model.Menu) error {
	return d.db.Save(menu).Error
}

func (d *MenuDAO) Delete(id uint) error {
	return d.db.Delete(&model.Menu{}, id).Error
}

func (d *MenuDAO) ListAll() ([]model.Menu, error) {
	var menus []model.Menu
	err := d.db.Where("status = 1").Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}
