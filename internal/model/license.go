package model

import "time"

// License 授权码
type License struct {
	BaseModel
	LicenseKey  string     `gorm:"size:1000;not null" json:"license_key"`
	MachineID   string     `gorm:"size:128;not null;index" json:"machine_id"`
	LicenseType string     `gorm:"size:20;not null;comment:monthly/yearly/permanent" json:"license_type"`
	ExpiresAt   *time.Time `json:"expires_at"`
	BoundAt     *time.Time `json:"bound_at"`
	Status      int8       `gorm:"default:0;comment:0-未激活 1-已激活 2-已过期" json:"status"`
}

func (License) TableName() string { return "licenses" }
