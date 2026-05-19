package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

// Config 采集器配置
type Config struct {
	APIServer   string `mapstructure:"api_server"`
	APIKey      string `mapstructure:"api_key"`
	AgentID     string `mapstructure:"agent_id"`
	BatchSize   int    `mapstructure:"batch_size"`
	FlushSec    int    `mapstructure:"flush_seconds"`
	AdminServer string `mapstructure:"admin_server"` // 管理后台地址，拉取任务和上报偏移量
}

// CollectTask 从管理后台拉取的采集任务
type CollectTask struct {
	ID               uint   `json:"id"`
	AgentID          uint   `json:"agent_id"`
	StoreID          uint   `json:"store_id"`
	StoreName        string `json:"store_name"`
	LogPathPattern   string `json:"log_path_pattern"`
	MultilinePattern string `json:"multiline_pattern"`
	ParseMode        string `json:"parse_mode"`
	ParseConfig      string `json:"parse_config"`
}

// FileOffset 文件偏移量
type FileOffset struct {
	TaskID    uint   `json:"task_id"`
	FilePath  string `json:"file_path"`
	FileInode uint64 `json:"file_inode"`
	Offset    int64  `json:"offset"`
}

var (
	cfg      Config
	hostname string
	buffer   []map[string]interface{}
	bufMu    sync.Mutex
	offsets  = map[string]*FileOffset{} // file_path -> offset
	offMu    sync.Mutex
)

func main() {
	configFile := flag.String("config", "configs/logcollect.yaml", "config file path")
	flag.Parse()

	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("read config failed: %v\n", err)
		os.Exit(1)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		fmt.Printf("unmarshal config failed: %v\n", err)
		os.Exit(1)
	}
	if cfg.BatchSize == 0 {
		cfg.BatchSize = 50
	}
	if cfg.FlushSec == 0 {
		cfg.FlushSec = 5
	}
	if cfg.AgentID == "" {
		cfg.AgentID = "agent-001"
	}

	hostname, _ = os.Hostname()
	fmt.Printf("LogCollect agent started, hostname: %s, agent_id: %s\n", hostname, cfg.AgentID)

	// 1. 从管理后台拉取采集任务
	tasks := fetchTasks()
	if len(tasks) == 0 {
		fmt.Println("No collect tasks found, waiting...")
	} else {
		fmt.Printf("Loaded %d collect tasks\n", len(tasks))
	}

	// 2. 从管理后台拉取偏移量
	fetchOffsets(tasks)

	// 3. 启动文件监控
	done := make(chan struct{})
	var wg sync.WaitGroup
	for _, task := range tasks {
		files, _ := filepath.Glob(task.LogPathPattern)
		for _, file := range files {
			wg.Add(1)
			go func(f string, t CollectTask) {
				defer wg.Done()
				tailFile(f, t, done)
			}(file, task)
		}
	}

	// 4. 定时flush
	go flushLoop(done)

	// 5. 定时上报偏移量（每10秒）
	go offsetReportLoop(done)

	// 6. 心跳上报
	go heartbeatLoop(done)

	// 7. 定期重新拉取任务（每60秒检查新任务）
	go taskRefreshLoop(tasks, done, &wg)

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down logcollect...")
	close(done)
	wg.Wait()
	flushBuffer()
	reportOffsets() // 最后上报一次偏移量
}

// fetchTasks 从管理后台拉取采集任务
func fetchTasks() []CollectTask {
	url := cfg.AdminServer + "/api/v1/agents/tasks?hostname=" + hostname
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("fetch tasks failed: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			List []CollectTask `json:"list"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Data.List
}

// fetchOffsets 拉取偏移量
func fetchOffsets(tasks []CollectTask) {
	url := fmt.Sprintf("%s/api/v1/agents/offsets?agent_id=%s", cfg.AdminServer, cfg.AgentID)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			List []FileOffset `json:"list"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	offMu.Lock()
	for _, o := range result.Data.List {
		offsets[o.FilePath] = &FileOffset{
			TaskID:    o.TaskID,
			FilePath:  o.FilePath,
			FileInode: o.FileInode,
			Offset:    o.Offset,
		}
	}
	offMu.Unlock()

	fmt.Printf("Loaded %d file offsets\n", len(result.Data.List))
}

// tailFile 追踪文件变化（支持断点续传）
func tailFile(filePath string, task CollectTask, done chan struct{}) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s, err: %v\n", filePath, err)
		return
	}
	defer file.Close()

	// 获取文件inode
	fi, _ := file.Stat()
	var inode uint64
	if si, ok := fi.Sys().(*syscall.Stat_t); ok {
		inode = si.Ino
	}

	// 检查是否有记录的偏移量
	offMu.Lock()
	savedOff, exists := offsets[filePath]
	offMu.Unlock()

	if exists && savedOff.FileInode == inode && savedOff.Offset > 0 {
		// 断点续传：从上次偏移量继续
		file.Seek(savedOff.Offset, io.SeekStart)
		fmt.Printf("Resuming %s from offset %d\n", filePath, savedOff.Offset)
	} else {
		// 新文件：从末尾开始
		file.Seek(0, io.SeekEnd)
		fmt.Printf("Tailing new file: %s\n", filePath)
	}

	buf := make([]byte, 4096)
	var lineBuf []byte

	for {
		select {
		case <-done:
			return
		default:
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				return
			}
			if n > 0 {
				lineBuf = append(lineBuf, buf[:n]...)
				// 按行分割
				for {
					idx := bytes.IndexByte(lineBuf, '\n')
					if idx < 0 {
						break
					}
					line := string(lineBuf[:idx])
					lineBuf = lineBuf[idx+1:]
					if line != "" {
						addToBuffer(filePath, task.StoreName, line)
					}
				}
				// 更新偏移量
				currentOff, _ := file.Seek(0, io.SeekCurrent)
				offMu.Lock()
				offsets[filePath] = &FileOffset{
					TaskID:    task.ID,
					FilePath:  filePath,
					FileInode: inode,
					Offset:    currentOff,
				}
				offMu.Unlock()
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func addToBuffer(filePath, storeName, message string) {
	bufMu.Lock()
	defer bufMu.Unlock()

	entry := map[string]interface{}{
		"message":    message,
		"file":       filePath,
		"hostname":   hostname,
		"timestamp":  time.Now().UnixMilli(),
		"store_name": storeName,
	}
	buffer = append(buffer, entry)

	if len(buffer) >= cfg.BatchSize {
		go flushBuffer()
	}
}

func flushLoop(done chan struct{}) {
	ticker := time.NewTicker(time.Duration(cfg.FlushSec) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			flushBuffer()
		}
	}
}

func flushBuffer() {
	bufMu.Lock()
	if len(buffer) == 0 {
		bufMu.Unlock()
		return
	}
	batch := buffer
	buffer = nil
	bufMu.Unlock()

	pushLogs(batch)
}

func pushLogs(logs []map[string]interface{}) {
	payload := map[string]interface{}{
		"api_key": cfg.APIKey,
		"logs":    logs,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	url := cfg.APIServer + "/api/v1/log/push"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("push logs failed: %v\n", err)
		return
	}
	resp.Body.Close()
}

// offsetReportLoop 定时上报偏移量
func offsetReportLoop(done chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			reportOffsets()
		}
	}
}

func reportOffsets() {
	offMu.Lock()
	var offsetList []FileOffset
	for _, o := range offsets {
		offsetList = append(offsetList, FileOffset{
			TaskID:    o.TaskID,
			FilePath:  o.FilePath,
			FileInode: o.FileInode,
			Offset:    o.Offset,
		})
	}
	offMu.Unlock()

	if len(offsetList) == 0 {
		return
	}

	payload := map[string]interface{}{
		"agent_id": cfg.AgentID,
		"offsets":  offsetList,
	}
	data, _ := json.Marshal(payload)
	url := cfg.AdminServer + "/api/v1/agents/offsets"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("report offsets failed: %v\n", err)
		return
	}
	resp.Body.Close()
}

func heartbeatLoop(done chan struct{}) {
	if cfg.AdminServer == "" {
		return
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			sendHeartbeat()
		}
	}
}

func sendHeartbeat() {
	payload := map[string]interface{}{
		"agent_id": cfg.AgentID,
		"hostname": hostname,
		"status":   "online",
		"time":     time.Now().Unix(),
	}
	data, _ := json.Marshal(payload)
	url := cfg.AdminServer + "/api/v1/agents/heartbeat"
	http.Post(url, "application/json", bytes.NewReader(data))
}

// taskRefreshLoop 定期检查新任务
func taskRefreshLoop(currentTasks []CollectTask, done chan struct{}, wg *sync.WaitGroup) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			newTasks := fetchTasks()
			// 找出新增的任务
			existingPaths := map[string]bool{}
			for _, t := range currentTasks {
				existingPaths[fmt.Sprintf("%d:%s", t.ID, t.LogPathPattern)] = true
			}
			for _, t := range newTasks {
				key := fmt.Sprintf("%d:%s", t.ID, t.LogPathPattern)
				if !existingPaths[key] {
					fmt.Printf("New task detected: %s -> %s\n", t.StoreName, t.LogPathPattern)
					files, _ := filepath.Glob(t.LogPathPattern)
					for _, file := range files {
						wg.Add(1)
						go func(f string, task CollectTask) {
							defer wg.Done()
							tailFile(f, task, done)
						}(file, t)
					}
					currentTasks = append(currentTasks, t)
				}
			}
		}
	}
}
