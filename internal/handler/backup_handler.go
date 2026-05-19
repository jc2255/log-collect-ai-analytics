package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// BackupHandler 备份快照管理
type BackupHandler struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewBackupHandler(db *gorm.DB, es *elastic.Client) *BackupHandler {
	return &BackupHandler{DB: db, ES: es}
}

// StartSnapshotSync 启动定时同步ES快照到数据库（每5分钟）
func (h *BackupHandler) StartSnapshotSync() {
	// 启动时立即同步一次
	h.syncSnapshots()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			h.syncSnapshots()
		}
	}()
}

// syncSnapshots 从ES拉取快照数据并同步入库
func (h *BackupHandler) syncSnapshots() {
	if h.ES == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	repos, err := h.ES.SnapshotGetRepository("_all").Do(ctx)
	if err != nil {
		return
	}

	for repoName := range repos {
		snapResp, err := h.ES.SnapshotGet(repoName).Snapshot("_all").Do(ctx)
		if err != nil {
			continue
		}
		for _, snap := range snapResp.Snapshots {
			indicesJSON, _ := json.Marshal(snap.Indices)
			var startTime, endTime string
			if !snap.StartTime.IsZero() {
				startTime = snap.StartTime.Format(time.RFC3339)
			}
			if !snap.EndTime.IsZero() {
				endTime = snap.EndTime.Format(time.RFC3339)
			}

			record := model.SnapshotRecord{
				SnapshotName: snap.Snapshot,
				UUID:         snap.UUID,
				State:        snap.State,
				Repository:   repoName,
				Indices:      string(indicesJSON),
				StartTime:    startTime,
				EndTime:      endTime,
				DurationMs:   snap.DurationInMillis,
			}
			// upsert: 按 snapshot_name 唯一索引，存在则更新
			var existing model.SnapshotRecord
			if h.DB.Where("snapshot_name = ?", snap.Snapshot).First(&existing).Error == nil {
				h.DB.Model(&existing).Updates(map[string]interface{}{
					"uuid":        record.UUID,
					"state":       record.State,
					"repository":  record.Repository,
					"indices":     record.Indices,
					"start_time":  record.StartTime,
					"end_time":    record.EndTime,
					"duration_ms": record.DurationMs,
				})
			} else {
				h.DB.Create(&record)
			}
		}
	}

	// 清理ES中已不存在但数据库中仍有的快照记录
	var dbRecords []model.SnapshotRecord
	h.DB.Find(&dbRecords)
	for _, rec := range dbRecords {
		found := false
		for repoName := range repos {
			snapResp, _ := h.ES.SnapshotGet(repoName).Snapshot(rec.SnapshotName).Do(ctx)
			if len(snapResp.Snapshots) > 0 {
				found = true
				break
			}
		}
		if !found {
			h.DB.Delete(&rec)
		}
	}
}

// ListSnapshots 分页查询快照列表（从数据库）
func (h *BackupHandler) ListSnapshots(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	repository := c.Query("repository")
	state := c.Query("state")

	query := h.DB.Model(&model.SnapshotRecord{})
	if repository != "" {
		query = query.Where("repository = ?", repository)
	}
	if state != "" {
		query = query.Where("state = ?", state)
	}

	var total int64
	query.Count(&total)

	var records []model.SnapshotRecord
	query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records)

	// 解析 indices JSON 为数组，方便前端展示
	type snapshotVO struct {
		model.SnapshotRecord
		IndicesList []string `json:"indices_list"`
	}
	var list []snapshotVO
	for _, r := range records {
		var indices []string
		if r.Indices != "" {
			json.Unmarshal([]byte(r.Indices), &indices)
		}
		list = append(list, snapshotVO{
			SnapshotRecord: r,
			IndicesList:    indices,
		})
	}
	if list == nil {
		list = []snapshotVO{}
	}

	response.Success(c, gin.H{"list": list, "total": total})
}

// DeleteSnapshot 删除快照
func (h *BackupHandler) DeleteSnapshot(c *gin.Context) {
	name := c.Param("name")
	repo := c.DefaultQuery("repo", "lca_backup")

	ctx := context.Background()
	_, err := h.ES.SnapshotDelete(repo, name).Do(ctx)
	if err != nil {
		response.Error(c, response.ErrorCode, fmt.Sprintf("删除失败: %v", err))
		return
	}
	// 同步删除数据库记录
	h.DB.Where("snapshot_name = ?", name).Delete(&model.SnapshotRecord{})
	response.Success(c, nil)
}

// RestoreSnapshot 恢复快照
func (h *BackupHandler) RestoreSnapshot(c *gin.Context) {
	name := c.Param("name")
	repo := c.DefaultQuery("repo", "lca_backup")

	ctx := context.Background()
	body := map[string]interface{}{}
	_, err := h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method: "POST",
		Path:   fmt.Sprintf("/_snapshot/%s/%s/_restore", repo, name),
		Body:   body,
	})
	if err != nil {
		response.Error(c, response.ErrorCode, fmt.Sprintf("恢复失败: %v", err))
		return
	}
	response.Success(c, nil)
}

// SLMHandler 备份策略管理
type SLMHandler struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewSLMHandler(db *gorm.DB, es *elastic.Client) *SLMHandler {
	return &SLMHandler{DB: db, ES: es}
}

func (h *SLMHandler) List(c *gin.Context) {
	var policies []model.SLMPolicy
	h.DB.Order("id desc").Find(&policies)
	response.Success(c, gin.H{"list": policies, "total": len(policies)})
}

func (h *SLMHandler) Create(c *gin.Context) {
	var req struct {
		Name           string `json:"name"`
		LogStore       string `json:"log_store"`
		Frequency      string `json:"frequency"`
		RetentionDays  int    `json:"retention_days"`
		MinCount       int    `json:"min_count"`
		MaxCount       int    `json:"max_count"`
		CronExpression string `json:"cron_expression"`
		Repository     string `json:"repository"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}
	policy := model.SLMPolicy{
		Name: req.Name, LogStore: req.LogStore, Frequency: req.Frequency,
		RetentionDays: req.RetentionDays, MinCount: req.MinCount,
		MaxCount: req.MaxCount, CronExpression: req.CronExpression,
		Repository: req.Repository, Status: 1,
	}
	if policy.Repository == "" {
		policy.Repository = "lca_backup"
	}
	// 在ES中创建SLM策略
	if h.ES != nil {
		h.createESSLMPolicy(policy)
	}
	if err := h.DB.Create(&policy).Error; err != nil {
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, policy)
}

func (h *SLMHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.SLMPolicy
	c.ShouldBindJSON(&req)
	h.DB.Model(&model.SLMPolicy{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name": req.Name, "log_store": req.LogStore, "frequency": req.Frequency,
		"retention_days": req.RetentionDays, "min_count": req.MinCount,
		"max_count": req.MaxCount, "cron_expression": req.CronExpression,
	})
	response.Success(c, nil)
}

func (h *SLMHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy model.SLMPolicy
	if h.DB.First(&policy, id).Error == nil {
		// 从ES删除SLM策略
		if h.ES != nil {
			ctx := context.Background()
			policyName := fmt.Sprintf("%s-slm-policy", policy.LogStore)
			h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
				Method: "DELETE",
				Path:   "/_slm/policy/" + policyName,
			})
		}
	}
	h.DB.Delete(&model.SLMPolicy{}, id)
	response.Success(c, nil)
}

// Execute 立即执行备份策略
func (h *SLMHandler) Execute(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var policy model.SLMPolicy
	if err := h.DB.First(&policy, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "策略不存在")
		return
	}

	if h.ES == nil {
		response.Error(c, response.ErrorCode, "ES未连接")
		return
	}

	ctx := context.Background()
	policyName := fmt.Sprintf("%s-slm-policy", policy.LogStore)
	_, err := h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method: "POST",
		Path:   "/_slm/policy/" + policyName + "/_execute",
	})
	if err != nil {
		response.Error(c, response.ErrorCode, fmt.Sprintf("执行失败: %v", err))
		return
	}
	response.Success(c, nil)
}

func (h *SLMHandler) createESSLMPolicy(policy model.SLMPolicy) {
	ctx := context.Background()
	policyName := fmt.Sprintf("%s-slm-policy", policy.LogStore)

	schedule := "0 0 1 * * ?" // default daily
	switch policy.Frequency {
	case "every_day":
		schedule = "0 0 1 * * ?"
	case "every_week":
		schedule = "0 0 1 ? * SUN"
	case "every_month":
		schedule = "0 0 1 1 * ?"
	}

	body := map[string]interface{}{
		"schedule":   schedule,
		"name":       fmt.Sprintf("<%s-snap-{now/d}>", policy.LogStore),
		"repository": policy.Repository,
		"config": map[string]interface{}{
			"indices": []string{policy.LogStore + "*"},
		},
		"retention": map[string]interface{}{
			"expire_after": fmt.Sprintf("%dd", policy.RetentionDays),
			"min_count":    policy.MinCount,
			"max_count":    policy.MaxCount,
		},
	}

	bodyJSON, _ := json.Marshal(body)
	h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method: "PUT",
		Path:   "/_slm/policy/" + policyName,
		Body:   string(bodyJSON),
	})
}
