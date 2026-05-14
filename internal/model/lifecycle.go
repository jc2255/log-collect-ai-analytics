package model

// ILMPolicy 索引生命周期管理策略
type ILMPolicy struct {
	BaseModel
	TenantID   uint   `gorm:"not null;index" json:"tenant_id"`
	StoreID    uint   `gorm:"not null;index" json:"store_id"`
	Name       string `gorm:"size:100;not null" json:"name"`
	HotDays    int    `gorm:"default:7" json:"hot_days"`
	WarmDays   int    `gorm:"default:30" json:"warm_days"`
	ColdDays   int    `gorm:"default:90" json:"cold_days"`
	DeleteDays int    `gorm:"default:180" json:"delete_days"`
	HotConfig  string `gorm:"type:text;comment:hot阶段配置JSON" json:"hot_config"`
	WarmConfig string `gorm:"type:text;comment:warm阶段配置JSON" json:"warm_config"`
	ColdConfig string `gorm:"type:text;comment:cold阶段配置JSON" json:"cold_config"`
	Status     int8   `gorm:"default:1" json:"status"`
}

func (ILMPolicy) TableName() string { return "ilm_policies" }

// SLMPolicy 快照生命周期管理策略
type SLMPolicy struct {
	BaseModel
	TenantID       uint   `gorm:"not null;index" json:"tenant_id"`
	StoreID        uint   `gorm:"not null;index" json:"store_id"`
	Name           string `gorm:"size:100;not null" json:"name"`
	ScheduleCron   string `gorm:"size:50;not null;comment:快照cron表达式" json:"schedule_cron"`
	Repository     string `gorm:"size:200;not null;comment:快照仓库名" json:"repository"`
	RetentionCount int    `gorm:"default:30;comment:保留快照数量" json:"retention_count"`
	RetentionDays  int    `gorm:"default:90;comment:保留天数" json:"retention_days"`
	Status         int8   `gorm:"default:1" json:"status"`
}

func (SLMPolicy) TableName() string { return "slm_policies" }

// BackupRecord 备份记录
type BackupRecord struct {
	BaseModel
	SLMPolicyID  uint   `gorm:"not null;index" json:"slm_policy_id"`
	TenantID     uint   `gorm:"not null;index" json:"tenant_id"`
	SnapshotName string `gorm:"size:200;not null" json:"snapshot_name"`
	Indices      string `gorm:"type:text;comment:备份的索引列表" json:"indices"`
	Status       string `gorm:"size:20;default:in_progress;comment:in_progress/success/failed" json:"status"`
	OSSPath      string `gorm:"size:500" json:"oss_path"`
	SizeBytes    int64  `gorm:"default:0" json:"size_bytes"`
	StartedAt    *int64 `json:"started_at"`
	FinishedAt   *int64 `json:"finished_at"`
	ErrorMsg     string `gorm:"size:1000" json:"error_msg"`
}

func (BackupRecord) TableName() string { return "backup_records" }

// Agent 采集Agent
type Agent struct {
	BaseModel
	TenantID      uint   `gorm:"not null;index" json:"tenant_id"`
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
	TenantID         uint   `gorm:"not null;index" json:"tenant_id"`
	LogPathPattern   string `gorm:"size:500;not null;comment:日志路径模式" json:"log_path_pattern"`
	MultilinePattern string `gorm:"size:500;comment:多行合并正则" json:"multiline_pattern"`
	ParseMode        string `gorm:"size:20;default:json;comment:json/regex/delimiter" json:"parse_mode"`
	ParseConfig      string `gorm:"type:text;comment:解析配置JSON" json:"parse_config"`
	Status           int8   `gorm:"default:1;comment:1-启用 0-禁用" json:"status"`
}

func (CollectTask) TableName() string { return "collect_tasks" }

// AuditLog 操作审计日志
type AuditLog struct {
	BaseModel
	TenantID  uint   `gorm:"not null;index" json:"tenant_id"`
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	Username  string `gorm:"size:50" json:"username"`
	Action    string `gorm:"size:100;not null" json:"action"`
	Resource  string `gorm:"size:200" json:"resource"`
	Detail    string `gorm:"type:text" json:"detail"`
	IP        string `gorm:"size:50" json:"ip"`
	UserAgent string `gorm:"size:500" json:"user_agent"`
}

func (AuditLog) TableName() string { return "audit_logs" }
