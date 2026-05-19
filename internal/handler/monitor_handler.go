package handler

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// MonitorHandler 服务监控
type MonitorHandler struct{}

func NewMonitorHandler() *MonitorHandler { return &MonitorHandler{} }

func (h *MonitorHandler) ServerInfo(c *gin.Context) {
	hostname, _ := os.Hostname()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := gin.H{
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"hostname":     hostname,
		"cpu_num":      runtime.NumCPU(),
		"goroutines":   runtime.NumGoroutine(),
		"go_version":   runtime.Version(),
		"mem_alloc":    fmt.Sprintf("%.2f MB", float64(memStats.Alloc)/1024/1024),
		"mem_total":    fmt.Sprintf("%.2f MB", float64(memStats.TotalAlloc)/1024/1024),
		"mem_sys":      fmt.Sprintf("%.2f MB", float64(memStats.Sys)/1024/1024),
		"gc_num":       memStats.NumGC,
		"startup_time": startupTime.Format("2006-01-02 15:04:05"),
		"run_duration": time.Since(startupTime).String(),
	}
	response.Success(c, info)
}

var startupTime = time.Now()

// LoginLogHandler 登录日志
type LoginLogHandler struct{ DB *gorm.DB }

func NewLoginLogHandler(db *gorm.DB) *LoginLogHandler { return &LoginLogHandler{DB: db} }

func (h *LoginLogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	username := c.Query("username")
	ip := c.Query("ip")
	status := c.Query("status")

	query := h.DB.Model(&model.LoginLog{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if ip != "" {
		query = query.Where("ip LIKE ?", "%"+ip+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var logs []model.LoginLog
	query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)
	response.Success(c, gin.H{"list": logs, "total": total})
}

func (h *LoginLogHandler) Delete(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if len(req.IDs) > 0 {
		h.DB.Where("id IN ?", req.IDs).Delete(&model.LoginLog{})
	}
	response.Success(c, nil)
}

func (h *LoginLogHandler) Clean(c *gin.Context) {
	h.DB.Where("1 = 1").Delete(&model.LoginLog{})
	response.Success(c, nil)
}

// OperLogHandler 操作日志
type OperLogHandler struct{ DB *gorm.DB }

func NewOperLogHandler(db *gorm.DB) *OperLogHandler { return &OperLogHandler{DB: db} }

func (h *OperLogHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	username := c.Query("username")
	resource := c.Query("resource")

	query := h.DB.Model(&model.AuditLog{})
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if resource != "" {
		query = query.Where("resource LIKE ?", "%"+resource+"%")
	}

	var total int64
	query.Count(&total)

	var logs []model.AuditLog
	query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs)
	response.Success(c, gin.H{"list": logs, "total": total})
}

func (h *OperLogHandler) Delete(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	c.ShouldBindJSON(&req)
	if len(req.IDs) > 0 {
		h.DB.Where("id IN ?", req.IDs).Delete(&model.AuditLog{})
	}
	response.Success(c, nil)
}

// OnlineHandler 在线用户
type OnlineHandler struct{ DB *gorm.DB }

func NewOnlineHandler(db *gorm.DB) *OnlineHandler { return &OnlineHandler{DB: db} }

func (h *OnlineHandler) List(c *gin.Context) {
	// 简易实现：从最近的登录日志获取活跃用户
	var logs []model.LoginLog
	since := time.Now().Add(-24 * time.Hour)
	h.DB.Where("status = 1 AND created_at > ?", since).
		Order("id desc").Limit(100).Find(&logs)

	// 按用户去重
	seen := make(map[string]bool)
	var result []model.LoginLog
	for _, log := range logs {
		if !seen[log.Username] {
			seen[log.Username] = true
			result = append(result, log)
		}
	}
	response.Success(c, gin.H{"list": result, "total": len(result)})
}

func (h *OnlineHandler) ForceLogout(c *gin.Context) {
	// 实际实现需要配合Redis token黑名单
	_ = middleware.GetCurrentUserID(c)
	response.Success(c, nil)
}
