package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/logger"
)

// AIAlertConfig AI告警配置结构
type AIAlertConfig struct {
	ScanIntervalMinutes int             `json:"scan_interval_minutes"`
	ErrorThreshold      int             `json:"error_threshold"`
	Keywords            []string        `json:"keywords"`
	LLMProvider         string          `json:"llm_provider"`
	LLMModel            string          `json:"llm_model"`
	LLMAPIKey           string          `json:"llm_api_key"`
	LLMBaseURL          string          `json:"llm_base_url"`
	NotifyChannels      []NotifyChannel `json:"notify_channels"`
	SilenceMinutes      int             `json:"silence_minutes"`
}

// NotifyChannel 通知渠道配置
type NotifyChannel struct {
	Type       string `json:"type"`        // wecom/dingtalk/email/webhook
	WebhookURL string `json:"webhook_url"` // webhook地址
	To         string `json:"to"`          // 邮件收件人
	SMTPHost   string `json:"smtp_host"`   // SMTP服务器
	SMTPPort   int    `json:"smtp_port"`   // SMTP端口
	SMTPUser   string `json:"smtp_user"`   // SMTP用户名
	SMTPPass   string `json:"smtp_pass"`   // SMTP密码
}

// ScanResult 扫描结果
type ScanResult struct {
	Triggered     bool                     `json:"triggered"`
	ErrorCount    int64                    `json:"error_count"`
	KeywordHits   map[string]int64         `json:"keyword_hits"`
	SampleLogs    []map[string]interface{} `json:"sample_logs"`
	TriggerReason string                   `json:"trigger_reason"`
}

// AIAlertScanner ES规则初筛扫描器
type AIAlertScanner struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewAIAlertScanner(db *gorm.DB, es *elastic.Client) *AIAlertScanner {
	return &AIAlertScanner{DB: db, ES: es}
}

// ScanStore 扫描指定日志库
func (s *AIAlertScanner) ScanStore(store *model.LogStore, isTest bool) {
	if store.AIAlertConfig == "" {
		logger.Infof("[AIAlert] store %s has no config, skip", store.Name)
		return
	}

	var cfg AIAlertConfig
	if err := json.Unmarshal([]byte(store.AIAlertConfig), &cfg); err != nil {
		logger.Errorf("[AIAlert] parse config for store %s failed: %v", store.Name, err)
		return
	}

	// 测试模式：直接发送测试通知，不依赖ES数据
	if isTest {
		testAnalysis := &LLMAnalysisResult{
			Severity:   "info",
			Summary:    fmt.Sprintf("[测试] 日志库 %s 的AI智能告警测试通知", store.Name),
			Impact:     "这是一条测试消息，无实际影响",
			Suggestion: "如果您收到此消息，说明告警通知渠道配置正确",
		}
		notifier := NewAIAlertNotifier()
		notifier.Send(store.Name, testAnalysis, cfg.NotifyChannels)

		// 记录测试告警历史
		history := model.AlertHistory{
			RuleID:   0,
			Content:  fmt.Sprintf("[测试告警] 日志库: %s\n摘要: %s", store.Name, testAnalysis.Summary),
			Status:   "resolved",
			RuleName: "ai_alert_" + store.Name,
			Severity: "info",
		}
		s.DB.Create(&history)
		logger.Infof("[AIAlert] test notification sent for store %s", store.Name)
		return
	}

	if s.ES == nil {
		logger.Errorf("[AIAlert] ES client is nil, skip scan for store %s", store.Name)
		return
	}

	// 检查静默期
	var lastAlert model.AlertHistory
	err := s.DB.Where("rule_name = ? AND created_at > ?",
		"ai_alert_"+store.Name,
		time.Now().Add(-time.Duration(cfg.SilenceMinutes)*time.Minute),
	).Order("id desc").First(&lastAlert).Error
	if err == nil {
		logger.Infof("[AIAlert] store %s in silence period, skip", store.Name)
		return
	}

	// 执行扫描
	result := s.scan(store, &cfg)
	if !result.Triggered {
		logger.Infof("[AIAlert] store %s scan completed, no alert triggered", store.Name)
		return
	}

	// 调用大模型分析
	llmService := NewAIAlertLLM(cfg.LLMProvider, cfg.LLMBaseURL, cfg.LLMAPIKey, cfg.LLMModel)
	analysis, err := llmService.Analyze(store.Name, result)
	if err != nil {
		logger.Errorf("[AIAlert] LLM analysis failed for store %s: %v", store.Name, err)
		// 降级：使用规则结果直接告警
		analysis = &LLMAnalysisResult{
			Severity:   "warning",
			Summary:    result.TriggerReason,
			Impact:     "需要人工确认影响范围",
			Suggestion: "请检查相关服务日志",
		}
	}

	// 发送通知
	notifier := NewAIAlertNotifier()
	notifier.Send(store.Name, analysis, cfg.NotifyChannels)

	// 记录告警历史
	history := model.AlertHistory{
		RuleID:   0, // AI告警不关联具体rule
		Content:  fmt.Sprintf("触发原因: %s\n分析摘要: %s\n影响: %s\n建议: %s", result.TriggerReason, analysis.Summary, analysis.Impact, analysis.Suggestion),
		Status:   "firing",
		RuleName: "ai_alert_" + store.Name,
		Severity: analysis.Severity,
	}
	s.DB.Create(&history)

	logger.Infof("[AIAlert] alert triggered for store %s: %s", store.Name, result.TriggerReason)
}

// scan 执行ES扫描
func (s *AIAlertScanner) scan(store *model.LogStore, cfg *AIAlertConfig) *ScanResult {
	result := &ScanResult{
		KeywordHits: make(map[string]int64),
	}

	ctx := context.Background()
	indexPattern := store.Name + "*"
	scanWindow := time.Duration(cfg.ScanIntervalMinutes) * time.Minute
	now := time.Now()
	startTime := now.Add(-scanWindow).Format(time.RFC3339)
	endTime := now.Format(time.RFC3339)

	// 1. 检查ERROR/FATAL日志数量
	errorQuery := elastic.NewBoolQuery().
		Must(elastic.NewTermsQueryFromStrings("level", "ERROR", "FATAL", "error", "fatal")).
		Filter(elastic.NewRangeQuery("@timestamp").Gte(startTime).Lte(endTime))

	errorCount, err := s.ES.Count(indexPattern).Query(errorQuery).Do(ctx)
	if err != nil {
		logger.Errorf("[AIAlert] count errors failed for %s: %v", store.Name, err)
		errorCount = 0
	}
	result.ErrorCount = errorCount

	if errorCount >= int64(cfg.ErrorThreshold) {
		result.Triggered = true
		result.TriggerReason = fmt.Sprintf("ERROR/FATAL日志数量(%d)超过阈值(%d)", errorCount, cfg.ErrorThreshold)
	}

	// 2. 检查关键词命中
	for _, kw := range cfg.Keywords {
		kwQuery := elastic.NewBoolQuery().
			Must(elastic.NewQueryStringQuery(kw)).
			Filter(elastic.NewRangeQuery("@timestamp").Gte(startTime).Lte(endTime))

		count, err := s.ES.Count(indexPattern).Query(kwQuery).Do(ctx)
		if err != nil {
			continue
		}
		if count > 0 {
			result.KeywordHits[kw] = count
		}
		// 关键词命中超5条也触发
		if count >= 5 && !result.Triggered {
			result.Triggered = true
			result.TriggerReason = fmt.Sprintf("关键词[%s]命中%d次", kw, count)
		}
	}

	// 3. 如果触发了，取样本日志
	if result.Triggered {
		sampleQuery := elastic.NewBoolQuery().
			Filter(elastic.NewRangeQuery("@timestamp").Gte(startTime).Lte(endTime))

		// 优先取ERROR级别的日志
		sampleQuery.Should(
			elastic.NewTermsQueryFromStrings("level", "ERROR", "FATAL", "error", "fatal"),
		)

		searchResult, err := s.ES.Search().
			Index(indexPattern).
			Query(sampleQuery).
			Sort("@timestamp", false).
			Size(50).
			Do(ctx)
		if err == nil && searchResult.Hits != nil {
			for _, hit := range searchResult.Hits.Hits {
				var doc map[string]interface{}
				json.Unmarshal(hit.Source, &doc)
				result.SampleLogs = append(result.SampleLogs, doc)
			}
		}
	}

	return result
}
