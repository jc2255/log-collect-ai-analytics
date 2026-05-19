package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// LogStoreHandler 日志库管理
type LogStoreHandler struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewLogStoreHandler(db *gorm.DB, es *elastic.Client) *LogStoreHandler {
	return &LogStoreHandler{DB: db, ES: es}
}

// generateAPIKey 生成随机API Key
func generateAPIKey(name string) string {
	b := make([]byte, 16)
	rand.Read(b)
	return "ak_" + name + "_" + hex.EncodeToString(b)[:8]
}

func (h *LogStoreHandler) List(c *gin.Context) {
	var stores []model.LogStore
	h.DB.Order("id desc").Find(&stores)
	response.Success(c, gin.H{"list": stores, "total": len(stores)})
}

func (h *LogStoreHandler) Create(c *gin.Context) {
	var req struct {
		Name               string `json:"name" binding:"required"`
		Description        string `json:"description"`
		IndexPattern       string `json:"index_pattern"`
		APIKey             string `json:"api_key"`
		Compress           *bool  `json:"compress"`
		RollMaxDays        int    `json:"roll_max_days"`
		RollMaxSizeGB      int    `json:"roll_max_size_gb"`
		ColdDays           int    `json:"cold_days"`
		DeleteDays         int    `json:"delete_days"`
		OSSRepository      string `json:"oss_repository"`
		OSSEndpoint        string `json:"oss_endpoint"`
		OSSBucket          string `json:"oss_bucket"`
		OSSAccessKeyID     string `json:"oss_access_key_id"`
		OSSAccessKeySecret string `json:"oss_access_key_secret"`
		OSSPath            string `json:"oss_path"`
		OSSChunkSize       string `json:"oss_chunk_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误: "+err.Error())
		return
	}

	compress := true
	if req.Compress != nil {
		compress = *req.Compress
	}
	deleteDays := 90
	if req.DeleteDays > 0 {
		deleteDays = req.DeleteDays
	}
	apiKey := req.APIKey
	if apiKey == "" {
		apiKey = generateAPIKey(req.Name)
	}

	store := model.LogStore{
		Name:               req.Name,
		Description:        req.Description,
		IndexPattern:       req.IndexPattern,
		APIKey:             apiKey,
		KafkaTopic:         "lca_" + req.Name,
		Status:             1,
		Compress:           compress,
		RollMaxDays:        req.RollMaxDays,
		RollMaxSizeGB:      req.RollMaxSizeGB,
		ColdDays:           req.ColdDays,
		DeleteDays:         deleteDays,
		OSSRepository:      req.OSSRepository,
		OSSEndpoint:        req.OSSEndpoint,
		OSSBucket:          req.OSSBucket,
		OSSAccessKeyID:     req.OSSAccessKeyID,
		OSSAccessKeySecret: req.OSSAccessKeySecret,
		OSSPath:            req.OSSPath,
		OSSChunkSize:       req.OSSChunkSize,
	}
	if store.OSSChunkSize == "" {
		store.OSSChunkSize = "500mb"
	}

	// 在ES中创建索引模板和ILM策略
	if h.ES != nil {
		h.createESIndexPolicy(store)
	}

	if err := h.DB.Create(&store).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			response.Error(c, response.ErrorCode, "日志库名称已存在")
			return
		}
		response.Error(c, response.ErrorCode, "创建失败: "+err.Error())
		return
	}
	response.Success(c, store)
}

func (h *LogStoreHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Description        string `json:"description"`
		IndexPattern       string `json:"index_pattern"`
		APIKey             string `json:"api_key"`
		Compress           *bool  `json:"compress"`
		RollMaxDays        int    `json:"roll_max_days"`
		RollMaxSizeGB      int    `json:"roll_max_size_gb"`
		ColdDays           int    `json:"cold_days"`
		DeleteDays         int    `json:"delete_days"`
		OSSRepository      string `json:"oss_repository"`
		OSSEndpoint        string `json:"oss_endpoint"`
		OSSBucket          string `json:"oss_bucket"`
		OSSAccessKeyID     string `json:"oss_access_key_id"`
		OSSAccessKeySecret string `json:"oss_access_key_secret"`
		OSSPath            string `json:"oss_path"`
		OSSChunkSize       string `json:"oss_chunk_size"`
	}
	c.ShouldBindJSON(&req)
	updates := map[string]interface{}{
		"description":           req.Description,
		"index_pattern":         req.IndexPattern,
		"api_key":               req.APIKey,
		"roll_max_days":         req.RollMaxDays,
		"roll_max_size_gb":      req.RollMaxSizeGB,
		"cold_days":             req.ColdDays,
		"delete_days":           req.DeleteDays,
		"oss_repository":        req.OSSRepository,
		"oss_endpoint":          req.OSSEndpoint,
		"oss_bucket":            req.OSSBucket,
		"oss_access_key_id":     req.OSSAccessKeyID,
		"oss_access_key_secret": req.OSSAccessKeySecret,
		"oss_path":              req.OSSPath,
		"oss_chunk_size":        req.OSSChunkSize,
	}
	if req.Compress != nil {
		updates["compress"] = *req.Compress
	}
	h.DB.Model(&model.LogStore{}).Where("id = ?", id).Updates(updates)
	response.Success(c, nil)
}

func (h *LogStoreHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var store model.LogStore
	if err := h.DB.First(&store, id).Error; err != nil {
		response.Error(c, response.ErrorCode, "日志库不存在")
		return
	}
	h.DB.Delete(&store)
	response.Success(c, nil)
}

// ESLogHandler 日志查询
type ESLogHandler struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewESLogHandler(db *gorm.DB, es *elastic.Client) *ESLogHandler {
	return &ESLogHandler{DB: db, ES: es}
}

func (h *ESLogHandler) Search(c *gin.Context) {
	storeName := c.Query("store")
	keyword := c.Query("keyword")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if storeName == "" {
		response.Error(c, response.ErrorCode, "请选择日志库")
		return
	}

	if h.ES == nil {
		response.Error(c, response.ErrorCode, "ES未连接")
		return
	}

	// 构建查询
	indexPattern := storeName + "*"
	boolQuery := elastic.NewBoolQuery()

	if keyword != "" {
		boolQuery.Must(elastic.NewQueryStringQuery(keyword))
	}
	if startTime != "" && endTime != "" {
		boolQuery.Filter(elastic.NewRangeQuery("@timestamp").Gte(startTime).Lte(endTime))
	}

	ctx := context.Background()
	searchResult, err := h.ES.Search().
		Index(indexPattern).
		Query(boolQuery).
		Sort("@timestamp", false).
		From((page - 1) * pageSize).
		Size(pageSize).
		Do(ctx)
	if err != nil {
		response.Error(c, response.ErrorCode, fmt.Sprintf("查询失败: %v", err))
		return
	}

	var logs []map[string]interface{}
	for _, hit := range searchResult.Hits.Hits {
		var doc map[string]interface{}
		json.Unmarshal(hit.Source, &doc)
		doc["_id"] = hit.Id
		logs = append(logs, doc)
	}

	response.Success(c, gin.H{
		"list":  logs,
		"total": searchResult.TotalHits(),
	})
}

// Fields 获取索引字段映射（Kibana Discover字段列表）
func (h *ESLogHandler) Fields(c *gin.Context) {
	storeName := c.Query("store")
	if storeName == "" {
		response.Error(c, response.ErrorCode, "请选择日志库")
		return
	}
	if h.ES == nil {
		response.Success(c, gin.H{"fields": []interface{}{}})
		return
	}

	ctx := context.Background()
	indexPattern := storeName + "*"
	// 获取mapping
	resp, err := h.ES.GetMapping().Index(indexPattern).Do(ctx)
	if err != nil {
		response.Success(c, gin.H{"fields": []interface{}{}})
		return
	}

	fieldSet := map[string]string{} // field -> type
	for _, indexMap := range resp {
		im, _ := indexMap.(map[string]interface{})
		if m, ok := im["mappings"]; ok {
			mm, _ := m.(map[string]interface{})
			if props, ok := mm["properties"]; ok {
				walkFields("", props.(map[string]interface{}), fieldSet)
			}
		}
	}

	var fields []gin.H
	for name, typ := range fieldSet {
		fields = append(fields, gin.H{"name": name, "type": typ})
	}
	response.Success(c, gin.H{"fields": fields})
}

func walkFields(prefix string, props map[string]interface{}, result map[string]string) {
	for name, val := range props {
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}
		if m, ok := val.(map[string]interface{}); ok {
			if t, has := m["type"]; has {
				result[fullName] = fmt.Sprintf("%v", t)
			}
			if subProps, has := m["properties"]; has {
				walkFields(fullName, subProps.(map[string]interface{}), result)
			}
		}
	}
}

// Histogram 时间直方图（Kibana Discover上方图表）
func (h *ESLogHandler) Histogram(c *gin.Context) {
	storeName := c.Query("store")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")
	keyword := c.Query("keyword")

	if storeName == "" || h.ES == nil {
		response.Success(c, gin.H{"buckets": []interface{}{}})
		return
	}

	indexPattern := storeName + "*"
	boolQuery := elastic.NewBoolQuery()
	if keyword != "" {
		boolQuery.Must(elastic.NewQueryStringQuery(keyword))
	}
	if startTime != "" && endTime != "" {
		boolQuery.Filter(elastic.NewRangeQuery("@timestamp").Gte(startTime).Lte(endTime))
	} else {
		// 默认最近15分钟
		boolQuery.Filter(elastic.NewRangeQuery("@timestamp").Gte("now-15m").Lte("now"))
	}

	// 自动计算interval
	interval := "1m"
	if startTime != "" && endTime != "" {
		duration := parseDuration(startTime, endTime)
		if duration > 72*time.Hour {
			interval = "3h"
		} else if duration > 24*time.Hour {
			interval = "1h"
		} else if duration > 4*time.Hour {
			interval = "30m"
		} else if duration > time.Hour {
			interval = "5m"
		}
	}

	agg := elastic.NewDateHistogramAggregation().Field("@timestamp").CalendarInterval(interval)

	ctx := context.Background()
	searchResult, err := h.ES.Search().
		Index(indexPattern).
		Query(boolQuery).
		Size(0).
		Aggregation("histogram", agg).
		Do(ctx)
	if err != nil {
		response.Success(c, gin.H{"buckets": []interface{}{}})
		return
	}

	var buckets []gin.H
	if aggResult, found := searchResult.Aggregations.DateHistogram("histogram"); found {
		for _, b := range aggResult.Buckets {
			buckets = append(buckets, gin.H{
				"key":        b.Key,
				"key_string": b.KeyAsString,
				"doc_count":  b.DocCount,
			})
		}
	}

	if buckets == nil {
		buckets = []gin.H{}
	}
	response.Success(c, gin.H{"buckets": buckets})
}

func parseDuration(start, end string) time.Duration {
	layout := "2006-01-02T15:04:05"
	if len(start) < 19 || len(end) < 19 {
		return time.Hour // 相对时间格式(如 now-1h)默认1小时
	}
	t1, err1 := time.Parse(layout, start[:19])
	t2, err2 := time.Parse(layout, end[:19])
	if err1 != nil || err2 != nil {
		return time.Hour
	}
	return t2.Sub(t1)
}

// createESIndexPolicy 在ES中创建ILM策略和索引模板
func (h *LogStoreHandler) createESIndexPolicy(store model.LogStore) {
	if h.ES == nil {
		return
	}
	ctx := context.Background()
	topic := store.Name

	// 检测ES版本（OSS版本不支持ILM/X-Pack）
	isOSS := false
	infoRes, err := h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method: "GET",
		Path:   "/",
	})
	if err == nil && infoRes != nil && infoRes.Body != nil {
		var infoMap map[string]interface{}
		if json.Unmarshal(infoRes.Body, &infoMap) == nil {
			if ver, ok := infoMap["version"].(map[string]interface{}); ok {
				if flavor, ok := ver["build_flavor"].(string); ok && flavor == "oss" {
					isOSS = true
				}
			}
		}
	}

	// 1. 创建ILM策略（仅非OSS版本）
	if !isOSS {
		var phases map[string]interface{}

		// hot phase
		hot := map[string]interface{}{"min_age": "0ms"}
		hotActions := map[string]interface{}{"rollover": map[string]interface{}{}}
		if store.RollMaxDays > 0 {
			hotActions["rollover"].(map[string]interface{})["max_age"] = fmt.Sprintf("%dd", store.RollMaxDays)
		}
		if store.RollMaxSizeGB > 0 {
			hotActions["rollover"].(map[string]interface{})["max_size"] = fmt.Sprintf("%dgb", store.RollMaxSizeGB)
		}
		hot["actions"] = hotActions

		phases = map[string]interface{}{"hot": hot}

		// warm phase
		if store.ColdDays > 0 {
			warm := map[string]interface{}{
				"min_age": fmt.Sprintf("%dd", store.ColdDays),
				"actions": map[string]interface{}{
					"forcemerge": map[string]interface{}{"max_num_segments": 1},
					"shrink":     map[string]interface{}{"number_of_shards": 1},
				},
			}
			phases["warm"] = warm
		}

		// delete phase
		if store.DeleteDays > 0 {
			deletePhase := map[string]interface{}{
				"min_age": fmt.Sprintf("%dd", store.DeleteDays),
				"actions": map[string]interface{}{
					"delete": map[string]interface{}{},
				},
			}
			phases["delete"] = deletePhase
		}

		ilmBody := map[string]interface{}{
			"policy": map[string]interface{}{
				"phases": phases,
			},
		}
		ilmJSON, _ := json.Marshal(ilmBody)
		h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
			Method: "PUT",
			Path:   fmt.Sprintf("/_ilm/policy/%s-ilm-policy", topic),
			Body:   string(ilmJSON),
		})
	}

	// 2. 创建索引模板（OSS版本也支持）
	templateBody := map[string]interface{}{
		"index_patterns": []string{topic + "-*"},
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 1,
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"@timestamp": map[string]interface{}{"type": "date"},
					"message":    map[string]interface{}{"type": "text"},
					"level":      map[string]interface{}{"type": "keyword"},
				},
			},
		},
	}
	if store.Compress {
		templateBody["template"].(map[string]interface{})["settings"].(map[string]interface{})["codec"] = "best_compression"
	}
	// 非OSS版本添加ILM关联
	if !isOSS {
		templateBody["template"].(map[string]interface{})["settings"].(map[string]interface{})["lifecycle.name"] = topic + "-ilm-policy"
		templateBody["template"].(map[string]interface{})["settings"].(map[string]interface{})["lifecycle.rollover_alias"] = topic
	}
	templateJSON, _ := json.Marshal(templateBody)
	h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method: "PUT",
		Path:   fmt.Sprintf("/_index_template/%s-template", topic),
		Body:   string(templateJSON),
	})

	// 3. 如果配置了OSS，创建S3兼容的快照仓库
	// 注意：OSS版本不支持SLM，但支持手动快照；密钥需在 elasticsearch.yml + keystore 中配置
	if store.OSSRepository != "" && store.OSSBucket != "" {
		repoBody := map[string]interface{}{
			"type": "s3",
			"settings": map[string]interface{}{
				"bucket":    store.OSSBucket,
				"base_path": store.OSSPath,
				"client":    "default",
				"compress":  store.Compress,
			},
		}
		repoJSON, _ := json.Marshal(repoBody)
		_, err := h.ES.PerformRequest(ctx, elastic.PerformRequestOptions{
			Method: "PUT",
			Path:   fmt.Sprintf("/_snapshot/%s", store.OSSRepository),
			Body:   string(repoJSON),
		})
		if err != nil {
			fmt.Printf("create snapshot repo failed: %v\n", err)
		} else {
			fmt.Printf("snapshot repo %s created successfully\n", store.OSSRepository)
		}
	}
}

// DashboardHandler 首页统计
type DashboardHandler struct {
	DB *gorm.DB
	ES *elastic.Client
}

func NewDashboardHandler(db *gorm.DB, es *elastic.Client) *DashboardHandler {
	return &DashboardHandler{DB: db, ES: es}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	// 日志库数量
	var storeCount int64
	h.DB.Model(&model.LogStore{}).Count(&storeCount)

	// 告警数量
	var alertCount int64
	h.DB.Model(&model.AlertHistory{}).Count(&alertCount)

	// Agent数量
	var agentCount int64
	h.DB.Model(&model.Agent{}).Count(&agentCount)

	// 日志总量 & 各日志库文档数
	var totalDocs int64
	storeStats := make([]gin.H, 0)
	var todayDocs int64

	var stores []model.LogStore
	h.DB.Find(&stores)

	if h.ES != nil {
		ctx := context.Background()
		now := time.Now()
		loc := now.Location()
		todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Format(time.RFC3339)
		todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc).Format(time.RFC3339)

		for _, store := range stores {
			indexPattern := store.Name + "*"
			count, err := h.ES.Count(indexPattern).Do(ctx)
			if err != nil {
				count = 0
			}
			totalDocs += count

			// 只统计有数据的日志库
			if count > 0 {
				storeStats = append(storeStats, gin.H{"name": store.Name, "doc_count": count})
			}

			// 今日文档数（精确到当天范围）
			todayCount, _ := h.ES.Count(indexPattern).
				Query(elastic.NewRangeQuery("@timestamp").Gte(todayStart).Lte(todayEnd)).
				Do(ctx)
			todayDocs += todayCount
		}

		// 按文档数降序排序
		for i := 0; i < len(storeStats); i++ {
			for j := i + 1; j < len(storeStats); j++ {
				if storeStats[i]["doc_count"].(int64) < storeStats[j]["doc_count"].(int64) {
					storeStats[i], storeStats[j] = storeStats[j], storeStats[i]
				}
			}
		}
	}

	response.Success(c, gin.H{
		"logstore_count": storeCount,
		"total_docs":     totalDocs,
		"alert_count":    alertCount,
		"agent_count":    agentCount,
		"today_docs":     todayDocs,
		"logstore_stats": storeStats,
	})
}

// 确保 time 包被使用
var _ = time.Now
