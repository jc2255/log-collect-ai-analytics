package dao

import (
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"gorm.io/gorm"
)

// UserDAO 用户数据访问
type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) Create(user *model.User) error {
	return d.db.Create(user).Error
}

func (d *UserDAO) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := d.db.Preload("Roles").First(&user, id).Error
	return &user, err
}

func (d *UserDAO) GetByUsername(tenantID uint, username string) (*model.User, error) {
	var user model.User
	err := d.db.Where("tenant_id = ? AND username = ?", tenantID, username).
		Preload("Roles").First(&user).Error
	return &user, err
}

func (d *UserDAO) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := d.db.Where("username = ?", username).Preload("Roles").Preload("Tenant").First(&user).Error
	return &user, err
}

func (d *UserDAO) Update(user *model.User) error {
	return d.db.Save(user).Error
}

func (d *UserDAO) Delete(id uint) error {
	return d.db.Delete(&model.User{}, id).Error
}

func (d *UserDAO) List(tenantID uint, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	query := d.db.Model(&model.User{})
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	query.Count(&total)
	err := query.Preload("Roles").Offset((page - 1) * pageSize).Limit(pageSize).
		Order("id DESC").Find(&users).Error
	return users, total, err
}

func (d *UserDAO) AssignRoles(userID uint, roleIDs []uint) error {
	var user model.User
	user.ID = userID
	var roles []model.Role
	for _, id := range roleIDs {
		roles = append(roles, model.Role{BaseModel: model.BaseModel{ID: id}})
	}
	return d.db.Model(&user).Association("Roles").Replace(roles)
}
