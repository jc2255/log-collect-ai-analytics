package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

// SyslogConfig syslog配置
type SyslogConfig struct {
	UDPPort    int            `mapstructure:"udp_port"`
	TCPPort    int            `mapstructure:"tcp_port"`
	APIServer  string         `mapstructure:"api_server"`
	Routes     []RouteConfig  `mapstructure:"routes"`
	BatchSize  int            `mapstructure:"batch_size"`
	FlushSec   int            `mapstructure:"flush_seconds"`
}

type RouteConfig struct {
	Match  string `mapstructure:"match"`  // 匹配规则，如 hostname 或 facility
	APIKey string `mapstructure:"api_key"`
}

var (
	syslogCfg SyslogConfig
	logBuffer []map[string]interface{}
)

func main() {
	configFile := flag.String("config", "configs/syslog.yaml", "config file path")
	flag.Parse()

	viper.SetConfigFile(*configFile)
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("read config failed: %v\n", err)
		os.Exit(1)
	}
	if err := viper.Unmarshal(&syslogCfg); err != nil {
		fmt.Printf("unmarshal config failed: %v\n", err)
		os.Exit(1)
	}

	if syslogCfg.UDPPort == 0 {
		syslogCfg.UDPPort = 514
	}
	if syslogCfg.TCPPort == 0 {
		syslogCfg.TCPPort = 514
	}
	if syslogCfg.BatchSize == 0 {
		syslogCfg.BatchSize = 50
	}
	if syslogCfg.FlushSec == 0 {
		syslogCfg.FlushSec = 3
	}

	fmt.Printf("Syslog server starting UDP:%d TCP:%d\n", syslogCfg.UDPPort, syslogCfg.TCPPort)

	// 启动UDP监听
	go startUDP()
	// 启动TCP监听
	go startTCP()
	// 启动定时flush
	go flushLoop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down syslog server...")
}

func startUDP() {
	addr := fmt.Sprintf(":%d", syslogCfg.UDPPort)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Printf("UDP listen failed: %v\n", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 65535)
	for {
		n, remoteAddr, err := conn.ReadFrom(buf)
		if err != nil {
			continue
		}
		msg := string(buf[:n])
		processSyslog(msg, remoteAddr.String())
	}
}

func startTCP() {
	addr := fmt.Sprintf(":%d", syslogCfg.TCPPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("TCP listen failed: %v\n", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleTCPConn(conn)
	}
}

func handleTCPConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 65535)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		msg := string(buf[:n])
		// TCP可能多条消息以换行分隔
		for _, line := range strings.Split(msg, "\n") {
			if line = strings.TrimSpace(line); line != "" {
				processSyslog(line, conn.RemoteAddr().String())
			}
		}
	}
}

func processSyslog(raw, sourceAddr string) {
	// 简单解析syslog格式
	entry := map[string]interface{}{
		"message":    raw,
		"source_ip":  strings.Split(sourceAddr, ":")[0],
		"timestamp":  time.Now().UnixMilli(),
		"protocol":   "syslog",
	}

	// 路由匹配确定api_key
	apiKey := matchRoute(raw)
	if apiKey == "" && len(syslogCfg.Routes) > 0 {
		apiKey = syslogCfg.Routes[0].APIKey // 默认使用第一个
	}

	entry["_api_key"] = apiKey
	logBuffer = append(logBuffer, entry)

	if len(logBuffer) >= syslogCfg.BatchSize {
		flush()
	}
}

func matchRoute(msg string) string {
	for _, route := range syslogCfg.Routes {
		if strings.Contains(msg, route.Match) {
			return route.APIKey
		}
	}
	return ""
}

func flushLoop() {
	ticker := time.NewTicker(time.Duration(syslogCfg.FlushSec) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		flush()
	}
}

func flush() {
	if len(logBuffer) == 0 {
		return
	}

	// 按api_key分组
	groups := make(map[string][]map[string]interface{})
	for _, entry := range logBuffer {
		key, _ := entry["_api_key"].(string)
		delete(entry, "_api_key")
		groups[key] = append(groups[key], entry)
	}
	logBuffer = nil

	for apiKey, logs := range groups {
		pushToAPI(apiKey, logs)
	}
}

func pushToAPI(apiKey string, logs []map[string]interface{}) {
	payload := map[string]interface{}{
		"api_key": apiKey,
		"logs":    logs,
	}
	data, _ := json.Marshal(payload)

	url := syslogCfg.APIServer + "/api/v1/log/push"
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Printf("push syslog failed: %v\n", err)
		return
	}
	resp.Body.Close()
}
