package service

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/logger"
)

// AIAlertScheduler 告警扫描调度器
type AIAlertScheduler struct {
	DB      *gorm.DB
	ES      *elastic.Client
	scanner *AIAlertScanner
	timers  map[uint]*time.Ticker
	stopChs map[uint]chan struct{}
	mu      sync.Mutex
}

func NewAIAlertScheduler(db *gorm.DB, es *elastic.Client) *AIAlertScheduler {
	return &AIAlertScheduler{
		DB:      db,
		ES:      es,
		scanner: NewAIAlertScanner(db, es),
		timers:  make(map[uint]*time.Ticker),
		stopChs: make(map[uint]chan struct{}),
	}
}

// Start 启动调度器，加载所有启用AI告警的日志库
func (s *AIAlertScheduler) Start() {
	var stores []model.LogStore
	s.DB.Where("ai_alert_enabled = ?", true).Find(&stores)

	logger.Infof("[AIAlertScheduler] starting, found %d stores with AI alert enabled", len(stores))

	for i := range stores {
		s.addStore(&stores[i])
	}

	// 启动定期刷新goroutine，每60秒检查一次是否有新启用/停用的日志库
	go s.refreshLoop()
}

// Stop 停止所有定时任务
func (s *AIAlertScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, stopCh := range s.stopChs {
		close(stopCh)
		if ticker, ok := s.timers[id]; ok {
			ticker.Stop()
		}
		delete(s.timers, id)
		delete(s.stopChs, id)
	}
	logger.Infof("[AIAlertScheduler] stopped all tasks")
}

func (s *AIAlertScheduler) addStore(store *model.LogStore) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已有定时任务先停止
	if stopCh, ok := s.stopChs[store.ID]; ok {
		close(stopCh)
		if ticker, ok := s.timers[store.ID]; ok {
			ticker.Stop()
		}
	}

	// 解析扫描间隔
	interval := 5 * time.Minute // 默认5分钟
	if store.AIAlertConfig != "" {
		var cfg AIAlertConfig
		if err := json.Unmarshal([]byte(store.AIAlertConfig), &cfg); err == nil && cfg.ScanIntervalMinutes > 0 {
			interval = time.Duration(cfg.ScanIntervalMinutes) * time.Minute
		}
	}

	ticker := time.NewTicker(interval)
	stopCh := make(chan struct{})
	s.timers[store.ID] = ticker
	s.stopChs[store.ID] = stopCh

	storeID := store.ID
	storeName := store.Name

	go func() {
		logger.Infof("[AIAlertScheduler] started scan task for store %s (id=%d), interval=%v", storeName, storeID, interval)
		for {
			select {
			case <-ticker.C:
				// 重新从数据库加载最新配置
				var latestStore model.LogStore
				if err := s.DB.First(&latestStore, storeID).Error; err != nil {
					logger.Errorf("[AIAlertScheduler] load store %d failed: %v", storeID, err)
					continue
				}
				if !latestStore.AIAlertEnabled {
					logger.Infof("[AIAlertScheduler] store %s AI alert disabled, stopping", storeName)
					return
				}
				s.scanner.ScanStore(&latestStore, false)
			case <-stopCh:
				logger.Infof("[AIAlertScheduler] stopped scan task for store %s (id=%d)", storeName, storeID)
				return
			}
		}
	}()
}

func (s *AIAlertScheduler) removeStore(storeID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stopCh, ok := s.stopChs[storeID]; ok {
		close(stopCh)
		delete(s.stopChs, storeID)
	}
	if ticker, ok := s.timers[storeID]; ok {
		ticker.Stop()
		delete(s.timers, storeID)
	}
}

func (s *AIAlertScheduler) refreshLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var stores []model.LogStore
		s.DB.Where("ai_alert_enabled = ?", true).Find(&stores)

		s.mu.Lock()
		// 找出新启用的
		enabledIDs := make(map[uint]bool)
		for i := range stores {
			enabledIDs[stores[i].ID] = true
			if _, exists := s.timers[stores[i].ID]; !exists {
				s.mu.Unlock()
				s.addStore(&stores[i])
				s.mu.Lock()
			}
		}
		// 找出已停用的
		for id := range s.timers {
			if !enabledIDs[id] {
				s.mu.Unlock()
				s.removeStore(id)
				s.mu.Lock()
			}
		}
		s.mu.Unlock()
	}
}
