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

// dropColumnIfExists 安全删除列（兼容MySQL不支持DROP COLUMN IF EXISTS）
func dropColumnIfExists(tableName, columnName string) {
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?",
		tableName, columnName).Scan(&count)
	if count > 0 {
		db.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, columnName))
	}
}

// dropFKIfExists 安全删除外键
func dropFKIfExists(tableName, fkName string) {
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = ? AND constraint_name = ? AND constraint_type = 'FOREIGN KEY'",
		tableName, fkName).Scan(&count)
	if count > 0 {
		db.Exec(fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", tableName, fkName))
	}
}

// dropStaleColumns 自动删除表中不在模型中的旧列
func dropStaleColumns(dst ...interface{}) {
	for _, m := range dst {
		stmt := &gorm.Statement{DB: db}
		_ = stmt.Parse(m)
		tableName := stmt.Table

		// 获取模型中的列名集合
		modelCols := map[string]bool{}
		for _, field := range stmt.Schema.Fields {
			modelCols[field.DBName] = true
		}
		// 基础列不算旧列
		modelCols["id"] = true
		modelCols["created_at"] = true
		modelCols["updated_at"] = true
		modelCols["deleted_at"] = true

		// 查询数据库中实际存在的列
		type colInfo struct {
			ColumnName string
		}
		var dbCols []colInfo
		db.Raw("SELECT column_name AS column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&dbCols)

		for _, col := range dbCols {
			if !modelCols[col.ColumnName] {
				db.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, col.ColumnName))
			}
		}
	}
}

// AutoMigrate 自动迁移数据表
func AutoMigrate() error {
	// 删除可能存在的旧外键约束
	dropFKIfExists("users", "fk_users_dept")
	dropFKIfExists("users", "fk_users_post")
	dropFKIfExists("users", "users_dept_id_fkey")
	dropFKIfExists("users", "users_post_id_fkey")

	// 清理旧的多租户 tenant_id 列（GORM AutoMigrate只加列不删列）
	tenantTables := []string{"users", "roles", "departments", "posts", "menus",
		"log_stores", "alert_rules", "alert_actions", "alert_histories",
		"slm_policies", "agents", "collect_tasks", "audit_logs", "login_logs"}
	for _, table := range tenantTables {
		dropColumnIfExists(table, "tenant_id")
	}

	// 自动清理模型中不存在的旧列
	dropStaleColumns(
		&User{}, &Role{}, &Department{}, &Post{}, &Menu{},
		&LogStore{}, &AlertRule{}, &AlertAction{}, &AlertHistory{},
		&SLMPolicy{}, &Agent{}, &CollectTask{}, &FileOffset{}, &AuditLog{}, &LoginLog{},
		&SnapshotRecord{},
	)

	return db.AutoMigrate(
		&User{},
		&Role{},
		&Department{},
		&Post{},
		&Menu{},
		&LogStore{},
		&AlertRule{},
		&AlertAction{},
		&AlertHistory{},
		&SLMPolicy{},
		&Agent{},
		&CollectTask{},
		&FileOffset{},
		&AuditLog{},
		&LoginLog{},
		&License{},
		&SnapshotRecord{},
	)
}
