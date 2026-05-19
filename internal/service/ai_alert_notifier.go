package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"github.com/cj/log-collect-ai-analytics/internal/pkg/logger"
)

// AIAlertNotifier 多渠道通知服务
type AIAlertNotifier struct {
	client *http.Client
}

func NewAIAlertNotifier() *AIAlertNotifier {
	return &AIAlertNotifier{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Send 发送通知到所有配置的渠道
func (n *AIAlertNotifier) Send(storeName string, analysis *LLMAnalysisResult, channels []NotifyChannel) {
	for _, ch := range channels {
		switch ch.Type {
		case "wecom":
			n.sendWecom(storeName, analysis, ch.WebhookURL)
		case "dingtalk":
			n.sendDingtalk(storeName, analysis, ch.WebhookURL)
		case "email":
			n.sendEmail(storeName, analysis, ch)
		case "webhook":
			n.sendWebhook(storeName, analysis, ch.WebhookURL)
		default:
			logger.Errorf("[AIAlert] unknown notify channel type: %s", ch.Type)
		}
	}
}

// sendWecom 发送企业微信机器人消息
func (n *AIAlertNotifier) sendWecom(storeName string, analysis *LLMAnalysisResult, webhookURL string) {
	if webhookURL == "" {
		return
	}

	severityEmoji := map[string]string{"critical": "🔴", "warning": "🟡", "info": "🔵"}
	emoji := severityEmoji[analysis.Severity]
	if emoji == "" {
		emoji = "⚠️"
	}

	content := fmt.Sprintf(`%s **LCA AI智能告警**
> **日志库**: %s
> **严重程度**: %s %s
> **异常摘要**: %s
> **影响范围**: %s
> **建议措施**: %s
> **告警时间**: %s`, emoji, storeName, emoji, analysis.Severity, analysis.Summary, analysis.Impact, analysis.Suggestion, time.Now().Format("2006-01-02 15:04:05"))

	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": content,
		},
	}

	n.postJSON(webhookURL, body, "wecom")
}

// sendDingtalk 发送钉钉机器人消息
func (n *AIAlertNotifier) sendDingtalk(storeName string, analysis *LLMAnalysisResult, webhookURL string) {
	if webhookURL == "" {
		return
	}

	title := fmt.Sprintf("LCA AI告警 - %s", storeName)
	text := fmt.Sprintf(`### LCA AI智能告警

- **日志库**: %s
- **严重程度**: %s
- **异常摘要**: %s
- **影响范围**: %s
- **建议措施**: %s
- **告警时间**: %s`, storeName, analysis.Severity, analysis.Summary, analysis.Impact, analysis.Suggestion, time.Now().Format("2006-01-02 15:04:05"))

	body := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  text,
		},
	}

	n.postJSON(webhookURL, body, "dingtalk")
}

// sendEmail 发送邮件通知（SMTP），支持 SSL(465) 和 STARTTLS(587) 两种模式
func (n *AIAlertNotifier) sendEmail(storeName string, analysis *LLMAnalysisResult, ch NotifyChannel) {
	if ch.To == "" {
		logger.Errorf("[AIAlert] email 'to' is empty, skip")
		return
	}
	if ch.SMTPHost == "" || ch.SMTPUser == "" || ch.SMTPPass == "" {
		logger.Errorf("[AIAlert] email SMTP config incomplete (host=%s, user=%s), skip", ch.SMTPHost, ch.SMTPUser)
		return
	}

	port := ch.SMTPPort
	if port == 0 {
		port = 465 // 默认 SSL 模式
	}

	subject := fmt.Sprintf("LCA AI告警 - %s [%s]", storeName, analysis.Severity)
	body := fmt.Sprintf(`LCA AI智能告警通知

日志库: %s
严重程度: %s
异常摘要: %s
影响范围: %s
建议措施: %s
告警时间: %s

---
此邮件由 LCA 日志系统自动发送`, storeName, analysis.Severity, analysis.Summary, analysis.Impact, analysis.Suggestion, time.Now().Format("2006-01-02 15:04:05"))

	recipients := strings.Split(ch.To, ",")
	for i := range recipients {
		recipients[i] = strings.TrimSpace(recipients[i])
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		ch.SMTPUser, strings.Join(recipients, ","), subject, body)

	addr := fmt.Sprintf("%s:%d", ch.SMTPHost, port)
	auth := smtp.PlainAuth("", ch.SMTPUser, ch.SMTPPass, ch.SMTPHost)

	var err error
	if port == 465 {
		// SSL 直连模式（QQ邮箱、163邮箱等）
		tlsCfg := &tls.Config{ServerName: ch.SMTPHost}
		conn, tlsErr := tls.Dial("tcp", addr, tlsCfg)
		if tlsErr != nil {
			logger.Errorf("[AIAlert] send email TLS dial %s failed: %v", addr, tlsErr)
			return
		}
		defer conn.Close()

		c, smtpErr := smtp.NewClient(conn, ch.SMTPHost)
		if smtpErr != nil {
			logger.Errorf("[AIAlert] smtp NewClient failed: %v", smtpErr)
			return
		}
		defer c.Quit()

		if err = c.Auth(auth); err != nil {
			logger.Errorf("[AIAlert] smtp auth failed: %v", err)
			return
		}
		if err = c.Mail(ch.SMTPUser); err != nil {
			logger.Errorf("[AIAlert] smtp MAIL FROM failed: %v", err)
			return
		}
		for _, rcpt := range recipients {
			if err = c.Rcpt(rcpt); err != nil {
				logger.Errorf("[AIAlert] smtp RCPT TO %s failed: %v", rcpt, err)
				return
			}
		}
		w, wErr := c.Data()
		if wErr != nil {
			logger.Errorf("[AIAlert] smtp DATA failed: %v", wErr)
			return
		}
		if _, err = w.Write([]byte(msg)); err != nil {
			logger.Errorf("[AIAlert] smtp write message failed: %v", err)
			return
		}
		err = w.Close()
	} else {
		// STARTTLS 模式（587端口，Gmail 等）
		err = smtp.SendMail(addr, auth, ch.SMTPUser, recipients, []byte(msg))
	}

	if err != nil {
		logger.Errorf("[AIAlert] send email to %s failed: %v", ch.To, err)
	} else {
		logger.Infof("[AIAlert] email notification sent to %s", ch.To)
	}
}

// sendWebhook 发送自定义Webhook
func (n *AIAlertNotifier) sendWebhook(storeName string, analysis *LLMAnalysisResult, webhookURL string) {
	if webhookURL == "" {
		return
	}

	body := map[string]interface{}{
		"store_name": storeName,
		"severity":   analysis.Severity,
		"summary":    analysis.Summary,
		"impact":     analysis.Impact,
		"suggestion": analysis.Suggestion,
		"timestamp":  time.Now().Format(time.RFC3339),
		"source":     "lca_ai_alert",
	}

	n.postJSON(webhookURL, body, "webhook")
}

func (n *AIAlertNotifier) postJSON(url string, body interface{}, channelType string) {
	data, err := json.Marshal(body)
	if err != nil {
		logger.Errorf("[AIAlert] marshal %s body failed: %v", channelType, err)
		return
	}

	resp, err := n.client.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		logger.Errorf("[AIAlert] send %s notification failed: %v", channelType, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		logger.Errorf("[AIAlert] %s webhook returned status %d", channelType, resp.StatusCode)
	} else {
		logger.Infof("[AIAlert] %s notification sent successfully", channelType)
	}
}
