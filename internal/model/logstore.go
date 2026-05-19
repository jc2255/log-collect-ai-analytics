package model

// LogStore 日志库
type LogStore struct {
	BaseModel
	Name         string `gorm:"size:100;not null;uniqueIndex" json:"name"`
	Description  string `gorm:"size:500" json:"description"`
	IndexPattern string `gorm:"size:200;comment:ES索引模式" json:"index_pattern"`
	APIKey       string `gorm:"size:100;index;comment:日志推送API Key" json:"api_key"`
	KafkaTopic   string `gorm:"size:200;comment:Kafka topic名称" json:"kafka_topic"`
	Status       int8   `gorm:"default:1" json:"status"`

	// ILM 索引生命周期管理字段
	Compress      bool `gorm:"default:true;comment:是否压缩存储" json:"compress"`
	RollMaxDays   int  `gorm:"default:0;comment:滚动最大天数(0不限)" json:"roll_max_days"`
	RollMaxSizeGB int  `gorm:"default:0;comment:滚动最大容量GB(0不限)" json:"roll_max_size_gb"`
	ColdDays      int  `gorm:"default:0;comment:冷存储天数(0不限)" json:"cold_days"`
	DeleteDays    int  `gorm:"default:90;comment:删除数据天数" json:"delete_days"`

	// AI 智能告警
	AIAlertEnabled bool   `gorm:"default:false;comment:是否启用AI智能告警" json:"ai_alert_enabled"`
	AIAlertConfig  string `gorm:"type:text;comment:AI告警配置JSON" json:"ai_alert_config"`

	// OSS 备份配置
	OSSRepository      string `gorm:"size:200;comment:OSS仓库名称" json:"oss_repository"`
	OSSEndpoint        string `gorm:"size:300;comment:OSS Endpoint" json:"oss_endpoint"`
	OSSBucket          string `gorm:"size:200;comment:OSS Bucket名称" json:"oss_bucket"`
	OSSAccessKeyID     string `gorm:"size:200;comment:OSS AccessKeyID" json:"oss_access_key_id"`
	OSSAccessKeySecret string `gorm:"size:200;comment:OSS AccessKeySecret" json:"oss_access_key_secret"`
	OSSPath            string `gorm:"size:500;comment:OSS存储路径" json:"oss_path"`
	OSSChunkSize       string `gorm:"size:20;default:500mb;comment:分块大小" json:"oss_chunk_size"`
}

func (LogStore) TableName() string { return "log_stores" }

// AlertRule 告警规则
type AlertRule struct {
	BaseModel
	StoreID          uint   `gorm:"not null;index" json:"store_id"`
	Name             string `gorm:"size:100;not null" json:"name"`
	Description      string `gorm:"size:500" json:"description"`
	QueryCondition   string `gorm:"type:text;not null;comment:查询条件DSL" json:"query_condition"`
	TriggerCondition string `gorm:"type:text;not null;comment:触发条件JSON" json:"trigger_condition"`
	Severity         string `gorm:"size:20;default:warning;comment:critical/warning/info" json:"severity"`
	CronExpr         string `gorm:"size:50;not null;comment:执行cron表达式" json:"cron_expr"`
	SilenceMinutes   int    `gorm:"default:5" json:"silence_minutes"`
	Status           int8   `gorm:"default:1" json:"status"`

	Actions []AlertAction `gorm:"foreignKey:RuleID" json:"actions,omitempty"`
}

func (AlertRule) TableName() string { return "alert_rules" }

// AlertAction 告警动作
type AlertAction struct {
	BaseModel
	RuleID     uint   `gorm:"not null;index" json:"rule_id"`
	ActionType string `gorm:"size:50;not null;comment:webhook/email/wecom/dingtalk" json:"action_type"`
	Config     string `gorm:"type:text;not null;comment:动作配置JSON" json:"config"`
	Status     int8   `gorm:"default:1" json:"status"`
}

func (AlertAction) TableName() string { return "alert_actions" }

// AlertHistory 告警历史
type AlertHistory struct {
	BaseModel
	RuleID   uint   `gorm:"not null;index" json:"rule_id"`
	Content  string `gorm:"type:text" json:"content"`
	Status   string `gorm:"size:20;default:firing;comment:firing/resolved" json:"status"`
	RuleName string `gorm:"size:100" json:"rule_name"`
	Severity string `gorm:"size:20" json:"severity"`
}

func (AlertHistory) TableName() string { return "alert_histories" }
