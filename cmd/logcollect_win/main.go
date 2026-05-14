package main

// logcollect_win is the Windows version of the log collection agent.
// It shares the same core logic as logcollect but uses Windows-specific paths.
// Build with: GOOS=windows GOARCH=amd64 go build -o logcollect.exe ./cmd/logcollect_win/

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

type Config struct {
	APIServer    string        `mapstructure:"api_server"`
	APIKey       string        `mapstructure:"api_key"`
	BatchSize    int           `mapstructure:"batch_size"`
	FlushSec     int           `mapstructure:"flush_seconds"`
	Collectors   []CollectConf `mapstructure:"collectors"`
	HeartbeatURL string        `mapstructure:"heartbeat_url"`
	AgentID      string        `mapstructure:"agent_id"`
}

type CollectConf struct {
	Paths            []string `mapstructure:"paths"`
	MultilinePattern string   `mapstructure:"multiline_pattern"`
	ParseMode        string   `mapstructure:"parse_mode"`
}

var (
	cfg      Config
	hostname string
	buffer   []map[string]interface{}
	bufMu    sync.Mutex
)

func main() {
	configFile := flag.String("config", "config.yaml", "config file path")
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

	hostname, _ = os.Hostname()
	fmt.Printf("LogCollect Windows agent started, hostname: %s\n", hostname)

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

	go flushLoop(ctx)
	go heartbeatLoop(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down logcollect_win...")
	close(ctx)
	wg.Wait()
	flushBuffer()
}

func tailFile(filePath string, done chan struct{}) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s, err: %v\n", filePath, err)
		return
	}
	defer file.Close()
	file.Seek(0, io.SeekEnd)

	buf := make([]byte, 4096)
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
				for _, line := range bytes.Split(buf[:n], []byte("\r\n")) {
					if len(line) > 0 {
						addToBuffer(filePath, string(line))
					}
				}
			} else {
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
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
	data, _ := json.Marshal(payload)
	url := cfg.APIServer + "/api/v1/log/push"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("push logs failed: %v\n", err)
		return
	}
	resp.Body.Close()
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
			payload := map[string]interface{}{
				"agent_id": cfg.AgentID,
				"hostname": hostname,
				"os_type":  "windows",
				"status":   "online",
				"time":     time.Now().Unix(),
			}
			data, _ := json.Marshal(payload)
			http.Post(cfg.HeartbeatURL, "application/json", bytes.NewReader(data))
		}
	}
}
