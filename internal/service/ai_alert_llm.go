package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cj/log-collect-ai-analytics/internal/pkg/logger"
)

// LLMAnalysisResult 大模型分析结果
type LLMAnalysisResult struct {
	Severity   string `json:"severity"`   // critical/warning/info
	Summary    string `json:"summary"`    // 异常摘要
	Impact     string `json:"impact"`     // 影响范围
	Suggestion string `json:"suggestion"` // 建议措施
}

// AIAlertLLM 大模型分析服务
type AIAlertLLM struct {
	Provider string
	BaseURL  string
	APIKey   string
	Model    string
}

func NewAIAlertLLM(provider, baseURL, apiKey, model string) *AIAlertLLM {
	if baseURL == "" {
		switch provider {
		case "deepseek":
			baseURL = "https://api.deepseek.com/v1"
		case "qwen":
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		case "ollama":
			baseURL = "http://localhost:11434/v1"
		default:
			baseURL = "https://api.openai.com/v1"
		}
	}
	return &AIAlertLLM{Provider: provider, BaseURL: baseURL, APIKey: apiKey, Model: model}
}

// Analyze 使用大模型分析异常日志
func (l *AIAlertLLM) Analyze(storeName string, scanResult *ScanResult) (*LLMAnalysisResult, error) {
	// 构建日志摘要
	logSummary := l.buildLogSummary(scanResult)

	prompt := fmt.Sprintf(`你是一位资深的日志分析专家。以下是来自日志库"%s"的异常日志信息：

触发原因：%s
ERROR日志数量：%d
关键词命中情况：%v

日志样本（最多50条）：
%s

请分析这些日志的异常模式，返回以下JSON格式的分析结果（不要返回其他内容，只返回纯JSON）：
{
  "severity": "critical或warning或info",
  "summary": "异常摘要（50字以内）",
  "impact": "影响范围描述（50字以内）",
  "suggestion": "建议措施（100字以内）"
}`, storeName, scanResult.TriggerReason, scanResult.ErrorCount, scanResult.KeywordHits, logSummary)

	// 调用OpenAI兼容API
	result, err := l.callLLMAPI(prompt)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (l *AIAlertLLM) buildLogSummary(scanResult *ScanResult) string {
	if len(scanResult.SampleLogs) == 0 {
		return "无样本日志"
	}

	var sb strings.Builder
	maxLogs := 20 // 只发送前20条避免超token
	if len(scanResult.SampleLogs) < maxLogs {
		maxLogs = len(scanResult.SampleLogs)
	}

	for i := 0; i < maxLogs; i++ {
		log := scanResult.SampleLogs[i]
		// 提取关键字段
		ts := fmt.Sprintf("%v", log["@timestamp"])
		level := fmt.Sprintf("%v", log["level"])
		msg := fmt.Sprintf("%v", log["message"])
		if msg == "<nil>" {
			msg = fmt.Sprintf("%v", log["msg"])
		}
		sb.WriteString(fmt.Sprintf("[%s] %s - %s\n", ts, level, msg))
	}

	return sb.String()
}

func (l *AIAlertLLM) callLLMAPI(prompt string) (*LLMAnalysisResult, error) {
	reqBody := map[string]interface{}{
		"model": l.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一位日志分析专家，只返回JSON格式的分析结果。"},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  500,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	url := strings.TrimSuffix(l.BaseURL, "/") + "/chat/completions"

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if l.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+l.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LLM API returned %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析OpenAI格式响应
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse LLM response failed: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("LLM returned no choices")
	}

	content := apiResp.Choices[0].Message.Content
	// 尝试从markdown代码块中提取JSON
	content = extractJSON(content)

	var result LLMAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		logger.Errorf("[AIAlert] parse LLM analysis result failed: %v, raw: %s", err, content)
		// 降级处理
		return &LLMAnalysisResult{
			Severity:   "warning",
			Summary:    content,
			Impact:     "需人工确认",
			Suggestion: "请查看详细日志",
		}, nil
	}

	return &result, nil
}

// extractJSON 从可能包含markdown代码块的文本中提取JSON
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	// 移除 ```json ... ```
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) > 2 {
			// 去掉第一行和最后一行
			lines = lines[1 : len(lines)-1]
			if strings.TrimSpace(lines[len(lines)-1]) == "```" {
				lines = lines[:len(lines)-1]
			}
			s = strings.Join(lines, "\n")
		}
	}
	return strings.TrimSpace(s)
}
