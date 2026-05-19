package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
	"github.com/cj/log-collect-ai-analytics/internal/service"
)

// AIAlertHandler AI智能告警管理
type AIAlertHandler struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewAIAlertHandler(db *gorm.DB, es *elastic.Client) *AIAlertHandler {
	return &AIAlertHandler{DB: db, ES: es}
}

// Toggle 启用/关闭 AI 告警
func (h *AIAlertHandler) Toggle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrorCode, "无效的ID")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}

	var store model.LogStore
	if err := h.DB.First(&store, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "日志库不存在")
		return
	}

	h.DB.Model(&store).Update("ai_alert_enabled", req.Enabled)
	response.Success(c, gin.H{"ai_alert_enabled": req.Enabled})
}

// GetConfig 获取 AI 告警配置
func (h *AIAlertHandler) GetConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrorCode, "无效的ID")
		return
	}

	var store model.LogStore
	if err := h.DB.First(&store, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "日志库不存在")
		return
	}

	response.Success(c, gin.H{
		"ai_alert_enabled": store.AIAlertEnabled,
		"ai_alert_config":  store.AIAlertConfig,
	})
}

// UpdateConfig 更新 AI 告警配置
func (h *AIAlertHandler) UpdateConfig(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrorCode, "无效的ID")
		return
	}

	var req struct {
		Config string `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}

	var store model.LogStore
	if err := h.DB.First(&store, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "日志库不存在")
		return
	}

	h.DB.Model(&store).Update("ai_alert_config", req.Config)
	response.Success(c, gin.H{"message": "配置已更新"})
}

// Test 测试告警（手动触发一次扫描）
func (h *AIAlertHandler) Test(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, response.ErrorCode, "无效的ID")
		return
	}

	var store model.LogStore
	if err := h.DB.First(&store, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "日志库不存在")
		return
	}

	if store.AIAlertConfig == "" {
		response.Error(c, response.ErrorCode, "请先配置AI告警参数")
		return
	}

	// 触发一次异步扫描测试
	go func() {
		scanner := service.NewAIAlertScanner(h.DB, h.ES)
		scanner.ScanStore(&store, true)
	}()

	response.Success(c, gin.H{"message": "测试告警已触发，请检查通知渠道"})
}

// History 获取告警历史列表
func (h *AIAlertHandler) History(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	storeID := c.Query("store_id")

	query := h.DB.Model(&model.AlertHistory{}).Order("id desc")
	if storeID != "" {
		query = query.Where("rule_id IN (SELECT id FROM alert_rules WHERE store_id = ?)", storeID)
	}

	var total int64
	query.Count(&total)

	var list []model.AlertHistory
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&list)

	response.Success(c, gin.H{"list": list, "total": total})
}
