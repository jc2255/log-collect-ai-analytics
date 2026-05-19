package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// CollectHandler 采集任务管理
type CollectHandler struct {
	DB *gorm.DB
}

func NewCollectHandler(db *gorm.DB) *CollectHandler {
	return &CollectHandler{DB: db}
}

// ListTasks 采集任务列表
func (h *CollectHandler) ListTasks(c *gin.Context) {
	var tasks []model.CollectTask
	h.DB.Order("id desc").Find(&tasks)
	response.Success(c, gin.H{"list": tasks, "total": len(tasks)})
}

// CreateTask 创建采集任务
func (h *CollectHandler) CreateTask(c *gin.Context) {
	var req struct {
		AgentID          uint   `json:"agent_id"`
		StoreID          uint   `json:"store_id" binding:"required"`
		StoreName        string `json:"store_name" binding:"required"`
		LogPathPattern   string `json:"log_path_pattern" binding:"required"`
		MultilinePattern string `json:"multiline_pattern"`
		ParseMode        string `json:"parse_mode"`
		ParseConfig      string `json:"parse_config"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}
	parseMode := req.ParseMode
	if parseMode == "" {
		parseMode = "raw"
	}
	task := model.CollectTask{
		AgentID:          req.AgentID,
		StoreID:          req.StoreID,
		StoreName:        req.StoreName,
		LogPathPattern:   req.LogPathPattern,
		MultilinePattern: req.MultilinePattern,
		ParseMode:        parseMode,
		ParseConfig:      req.ParseConfig,
		Status:           1,
	}
	if err := h.DB.Create(&task).Error; err != nil {
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, task)
}

// UpdateTask 更新采集任务
func (h *CollectHandler) UpdateTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		LogPathPattern   string `json:"log_path_pattern"`
		MultilinePattern string `json:"multiline_pattern"`
		ParseMode        string `json:"parse_mode"`
		ParseConfig      string `json:"parse_config"`
		Status           *int8  `json:"status"`
	}
	c.ShouldBindJSON(&req)
	updates := map[string]interface{}{
		"log_path_pattern":  req.LogPathPattern,
		"multiline_pattern": req.MultilinePattern,
		"parse_mode":        req.ParseMode,
		"parse_config":      req.ParseConfig,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	h.DB.Model(&model.CollectTask{}).Where("id = ?", id).Updates(updates)
	response.Success(c, nil)
}

// DeleteTask 删除采集任务
func (h *CollectHandler) DeleteTask(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	h.DB.Delete(&model.CollectTask{}, id)
	// 同时删除该任务的偏移量记录
	h.DB.Where("task_id = ?", id).Delete(&model.FileOffset{})
	response.Success(c, nil)
}

// ListAgents Agent列表
func (h *CollectHandler) ListAgents(c *gin.Context) {
	var agents []model.Agent
	h.DB.Order("id desc").Find(&agents)
	response.Success(c, gin.H{"list": agents, "total": len(agents)})
}

// DeleteAgent 删除Agent
func (h *CollectHandler) DeleteAgent(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	// 删除关联的偏移量记录
	h.DB.Where("agent_id = ?", id).Delete(&model.FileOffset{})
	h.DB.Delete(&model.Agent{}, id)
	response.Success(c, nil)
}

// Heartbeat Agent心跳上报
func (h *CollectHandler) Heartbeat(c *gin.Context) {
	var req struct {
		AgentID  string `json:"agent_id"`
		Hostname string `json:"hostname"`
		IP       string `json:"ip"`
		OSType   string `json:"os_type"`
		Version  string `json:"version"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}
	now := time.Now().Unix()
	var agent model.Agent
	result := h.DB.Where("hostname = ?", req.Hostname).First(&agent)
	if result.Error == gorm.ErrRecordNotFound {
		agent = model.Agent{
			Hostname:      req.Hostname,
			IP:            req.IP,
			OSType:        req.OSType,
			Version:       req.Version,
			Status:        "online",
			LastHeartbeat: &now,
		}
		h.DB.Create(&agent)
	} else {
		h.DB.Model(&agent).Updates(map[string]interface{}{
			"ip":             req.IP,
			"os_type":        req.OSType,
			"version":        req.Version,
			"status":         "online",
			"last_heartbeat": now,
		})
	}
	response.Success(c, nil)
}

// GetTasksForAgent Agent拉取自己的采集任务
func (h *CollectHandler) GetTasksForAgent(c *gin.Context) {
	hostname := c.Query("hostname")
	var tasks []model.CollectTask
	if hostname != "" {
		// 查找该主机对应的Agent
		var agent model.Agent
		if err := h.DB.Where("hostname = ?", hostname).First(&agent).Error; err == nil {
			// 返回: 绑定到该Agent的任务 + 全局任务(agent_id=0)
			h.DB.Where("(agent_id = ? OR agent_id = 0) AND status = 1", agent.ID).Find(&tasks)
		} else {
			// Agent未注册，返回全局任务
			h.DB.Where("agent_id = 0 AND status = 1").Find(&tasks)
		}
	}
	// 兼容：如果没有匹配的任务，返回所有启用的任务
	if len(tasks) == 0 {
		h.DB.Where("status = 1").Find(&tasks)
	}
	response.Success(c, gin.H{"list": tasks})
}

// GetOffsets Agent拉取自己的偏移量
func (h *CollectHandler) GetOffsets(c *gin.Context) {
	agentID := c.Query("agent_id")
	taskID := c.Query("task_id")
	var offsets []model.FileOffset
	q := h.DB.Model(&model.FileOffset{})
	if agentID != "" {
		q = q.Where("agent_id = ?", agentID)
	}
	if taskID != "" {
		q = q.Where("task_id = ?", taskID)
	}
	q.Find(&offsets)
	response.Success(c, gin.H{"list": offsets})
}

// UpdateOffsets 批量更新偏移量
func (h *CollectHandler) UpdateOffsets(c *gin.Context) {
	var req struct {
		AgentID string `json:"agent_id"`
		Offsets []struct {
			TaskID    uint   `json:"task_id"`
			FilePath  string `json:"file_path"`
			FileInode uint64 `json:"file_inode"`
			Offset    int64  `json:"offset"`
		} `json:"offsets"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}
	for _, o := range req.Offsets {
		h.DB.Where(model.FileOffset{
			AgentID:  req.AgentID,
			TaskID:   o.TaskID,
			FilePath: o.FilePath,
		}).Assign(model.FileOffset{
			FileInode: o.FileInode,
			Offset:    o.Offset,
		}).FirstOrCreate(&model.FileOffset{})
	}
	response.Success(c, nil)
}
