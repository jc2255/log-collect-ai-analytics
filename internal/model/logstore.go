package model

// LogStore 日志库
type LogStore struct {
	BaseModel
	TenantID      uint   `gorm:"not null;index" json:"tenant_id"`
	Name          string `gorm:"size:100;not null" json:"name"`
	Code          string `gorm:"size:50;not null" json:"code"`
	Description   string `gorm:"size:500" json:"description"`
	IndexPattern  string `gorm:"size:200;comment:ES索引模式" json:"index_pattern"`
	RetentionDays int    `gorm:"default:30;comment:保留天数" json:"retention_days"`
	APIKey        string `gorm:"size:100;uniqueIndex;comment:日志推送API Key" json:"api_key"`
	KafkaTopic    string `gorm:"size:200;comment:Kafka topic名称" json:"kafka_topic"`
	Status        int8   `gorm:"default:1" json:"status"`

	Fields []LogStoreField `gorm:"foreignKey:StoreID" json:"fields,omitempty"`
}

func (LogStore) TableName() string { return "log_stores" }

// LogStoreField 日志库字段定义
type LogStoreField struct {
	BaseModel
	StoreID   uint   `gorm:"not null;index" json:"store_id"`
	FieldName string `gorm:"size:100;not null" json:"field_name"`
	FieldType string `gorm:"size:50;not null;comment:keyword/text/long/double/date/ip" json:"field_type"`
	IsIndexed bool   `gorm:"default:true" json:"is_indexed"`
	IsKeyword bool   `gorm:"default:false" json:"is_keyword"`
	Analyzer  string `gorm:"size:50;comment:分词器" json:"analyzer"`
	Sort      int    `gorm:"default:0" json:"sort"`
}

func (LogStoreField) TableName() string { return "log_store_fields" }

// AlertRule 告警规则
type AlertRule struct {
	BaseModel
	TenantID         uint   `gorm:"not null;index" json:"tenant_id"`
	StoreID          uint   `gorm:"not null;index" json:"store_id"`
	Name             string `gorm:"size:100;not null" json:"name"`
	Description      string `gorm:"size:500" json:"description"`
	QueryCondition   string `gorm:"type:text;not null;comment:查询条件DSL" json:"query_condition"`
	TriggerCondition string `gorm:"type:text;not null;comment:触发条件JSON" json:"trigger_condition"`
	Severity         string `gorm:"size:20;default:warning;comment:critical/warning/info" json:"severity"`
	CronExpr         string `gorm:"size:50;not null;comment:执行cron表达式" json:"cron_expr"`
	SilenceMinutes   int    `gorm:"default:5" json:"silence_minutes"`
	GroupBy          string `gorm:"size:200;comment:分组字段" json:"group_by"`
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
	GroupKey string `gorm:"size:200;comment:分组key" json:"group_key"`
	RuleName string `gorm:"size:100" json:"rule_name"`
	Severity string `gorm:"size:20" json:"severity"`
}

func (AlertHistory) TableName() string { return "alert_histories" }
