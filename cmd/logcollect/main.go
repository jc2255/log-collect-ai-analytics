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
	APIServer   string        `mapstructure:"api_server"`
	APIKey      string        `mapstructure:"api_key"`
	BatchSize   int           `mapstructure:"batch_size"`
	FlushSec    int           `mapstructure:"flush_seconds"`
	Collectors  []CollectConf `mapstructure:"collectors"`
	HeartbeatURL string       `mapstructure:"heartbeat_url"`
	AgentID     string        `mapstructure:"agent_id"`
}

type CollectConf struct {
	Paths            []string `mapstructure:"paths"`
	MultilinePattern string   `mapstructure:"multiline_pattern"`
	ParseMode        string   `mapstructure:"parse_mode"`
}

// LogEntry 日志条目
type LogEntry struct {
	Message   string `json:"message"`
	File      string `json:"file"`
	Hostname  string `json:"hostname"`
	Timestamp int64  `json:"timestamp"`
}

var (
	cfg      Config
	hostname string
	buffer   []map[string]interface{}
	bufMu    sync.Mutex
)

func main() {
	configFile := flag.String("config", "configs/logcollect.yaml", "config file path")
	flag.Parse()

	// 加载配置
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

	hostname, _ = os.Hostname()
	fmt.Printf("LogCollect agent started, hostname: %s\n", hostname)

	// 启动文件监控
	ctx := make(chan struct{})
	var wg sync.WaitGroup

	for _, collector := range cfg.Collectors {
		for _, pattern := range collector.Paths {
			files, _ := filepath.Glob(pattern)
			for _, file := range files {
				wg.Add(1)
				go func(f string) {
					defer wg.Done()
					tailFile(f, ctx)
				}(file)
			}
		}
	}

	// 启动定时flush
	go flushLoop(ctx)

	// 启动心跳上报
	go heartbeatLoop(ctx)

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down logcollect...")
	close(ctx)
	wg.Wait()

	// flush剩余数据
	flushBuffer()
}

// tailFile 追踪文件变化
func tailFile(filePath string, done chan struct{}) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s, err: %v\n", filePath, err)
		return
	}
	defer file.Close()

	// 定位到文件末尾
	file.Seek(0, io.SeekEnd)

	buf := make([]byte, 4096)
	for {
		select {
		case <-done:
			return
		default:
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				fmt.Printf("read file error: %v\n", err)
				return
			}
			if n > 0 {
				lines := splitLines(buf[:n])
				for _, line := range lines {
					if line == "" {
						continue
					}
					addToBuffer(filePath, line)
				}
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func splitLines(data []byte) []string {
	var lines []string
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	return lines
}

func addToBuffer(filePath, message string) {
	bufMu.Lock()
	defer bufMu.Unlock()

	entry := map[string]interface{}{
		"message":   message,
		"file":      filePath,
		"hostname":  hostname,
		"timestamp": time.Now().UnixMilli(),
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
		fmt.Printf("marshal logs failed: %v\n", err)
		return
	}

	url := cfg.APIServer + "/api/v1/log/push"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("push logs failed: %v\n", err)
		// TODO: 写入本地缓冲文件，待恢复后重试
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("push logs response error: %s\n", string(body))
	}
}

func heartbeatLoop(done chan struct{}) {
	if cfg.HeartbeatURL == "" {
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
	http.Post(cfg.HeartbeatURL, "application/json", bytes.NewReader(data))
}
