package model

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dsn string) error {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return fmt.Errorf("connect to mysql failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(10)

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// TenantDB 带租户隔离的数据库操作辅助
type TenantDB struct {
	TenantID     uint
	IsSuperAdmin bool
}

// DB 获取带租户条件的数据库实例
func (t *TenantDB) DB() *gorm.DB {
	if t.IsSuperAdmin && t.TenantID == 0 {
		return db
	}
	return db.Where("tenant_id = ?", t.TenantID)
}

// AutoMigrate 自动迁移数据表
func AutoMigrate() error {
	return db.AutoMigrate(
		&Tenant{},
		&User{},
		&Role{},
		&Permission{},
		&Menu{},
		&LogStore{},
		&LogStoreField{},
		&AlertRule{},
		&AlertAction{},
		&AlertHistory{},
		&ILMPolicy{},
		&SLMPolicy{},
		&BackupRecord{},
		&Agent{},
		&CollectTask{},
		&AuditLog{},
	)
}
