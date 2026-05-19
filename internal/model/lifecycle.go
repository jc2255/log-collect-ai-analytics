package model

import "time"

// SLMPolicy 备份策略
type SLMPolicy struct {
	BaseModel
	Name           string `gorm:"size:100;not null" json:"name"`
	LogStore       string `gorm:"size:100;not null;comment:日志库名称" json:"log_store"`
	Frequency      string `gorm:"size:20;not null;comment:every_day/every_week/every_month" json:"frequency"`
	RetentionDays  int    `gorm:"default:30" json:"retention_days"`
	MinCount       int    `gorm:"default:5" json:"min_count"`
	MaxCount       int    `gorm:"default:100" json:"max_count"`
	CronExpression string `gorm:"size:100" json:"cron_expression"`
	Repository     string `gorm:"size:200;default:lca_backup" json:"repository"`
	Status         int8   `gorm:"default:1" json:"status"`
}

func (SLMPolicy) TableName() string { return "slm_policies" }

// Agent 采集Agent
type Agent struct {
	BaseModel
	Hostname      string `gorm:"size:200;not null" json:"hostname"`
	IP            string `gorm:"size:50;not null" json:"ip"`
	OSType        string `gorm:"size:20;comment:linux/windows" json:"os_type"`
	Version       string `gorm:"size:50" json:"version"`
	Status        string `gorm:"size:20;default:offline;comment:online/offline" json:"status"`
	LastHeartbeat *int64 `json:"last_heartbeat"`
	Labels        string `gorm:"type:text;comment:标签JSON" json:"labels"`
}

func (Agent) TableName() string { return "agents" }

// CollectTask 采集任务
type CollectTask struct {
	BaseModel
	AgentID          uint   `gorm:"not null;index" json:"agent_id"`
	StoreID          uint   `gorm:"not null;index" json:"store_id"`
	StoreName        string `gorm:"size:100;not null;comment:日志库名称" json:"store_name"`
	LogPathPattern   string `gorm:"size:500;not null;comment:日志路径模式" json:"log_path_pattern"`
	MultilinePattern string `gorm:"size:500;comment:多行合并正则" json:"multiline_pattern"`
	ParseMode        string `gorm:"size:20;default:json;comment:json/regex/delimiter/raw" json:"parse_mode"`
	ParseConfig      string `gorm:"type:text;comment:解析配置JSON" json:"parse_config"`
	Status           int8   `gorm:"default:1;comment:1-启用 0-禁用" json:"status"`
}

func (CollectTask) TableName() string { return "collect_tasks" }

// FileOffset 文件采集偏移量记录（类似Filebeat registry）
type FileOffset struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	AgentID   string    `gorm:"size:100;not null;index" json:"agent_id"`
	TaskID    uint      `gorm:"not null;index" json:"task_id"`
	FilePath  string    `gorm:"size:500;not null;index" json:"file_path"`
	FileInode uint64    `gorm:"not null;comment:文件inode" json:"file_inode"`
	Offset    int64     `gorm:"not null;default:0;comment:采集偏移量" json:"offset"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (FileOffset) TableName() string { return "file_offsets" }

// AuditLog 操作审计日志
type AuditLog struct {
	BaseModel
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Username  string `gorm:"size:50" json:"username"`
	Action    string `gorm:"size:100;not null" json:"action"`
	Module    string `gorm:"size:100" json:"module"`
	Resource  string `gorm:"size:200" json:"resource"`
	Detail    string `gorm:"type:text" json:"detail"`
	IP        string `gorm:"size:50" json:"ip"`
	UserAgent string `gorm:"size:500" json:"user_agent"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// LoginLog 登录日志
type LoginLog struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	Username  string    `gorm:"size:50" json:"username"`
	IP        string    `gorm:"size:50" json:"ip"`
	Location  string    `gorm:"size:200" json:"location"`
	Browser   string    `gorm:"size:200" json:"browser"`
	OS        string    `gorm:"size:200" json:"os"`
	Status    int8      `gorm:"default:1;comment:1-成功 0-失败" json:"status"`
	Msg       string    `gorm:"size:500" json:"msg"`
	Module    string    `gorm:"size:50;default:后台" json:"module"`
	CreatedAt time.Time `json:"created_at"`
}

func (LoginLog) TableName() string { return "login_logs" }

// SnapshotRecord ES快照记录（定时从ES同步入库，支持分页查询）
type SnapshotRecord struct {
	BaseModel
	SnapshotName string `gorm:"size:200;not null;uniqueIndex" json:"snapshot_name"`
	UUID         string `gorm:"size:100" json:"uuid"`
	State        string `gorm:"size:50;not null;comment:SUCCESS/IN_PROGRESS/FAILED" json:"state"`
	Repository   string `gorm:"size:200;index" json:"repository"`
	Indices      string `gorm:"type:text;comment:索引列表JSON" json:"indices"`
	StartTime    string `gorm:"size:50" json:"start_time"`
	EndTime      string `gorm:"size:50" json:"end_time"`
	DurationMs   int64  `json:"duration_ms"`
}

func (SnapshotRecord) TableName() string { return "snapshot_records" }
