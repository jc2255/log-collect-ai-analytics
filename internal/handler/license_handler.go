package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/middleware"
	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/license"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/response"
)

// LicenseHandler 授权码管理
type LicenseHandler struct {
	DB        *gorm.DB
	Checker   *license.Verifier
	LCATopURL string // lca.top 远程验证地址，为空则降级本地验签
}

func NewLicenseHandler(db *gorm.DB, checker *license.Verifier, lcaTopURL string) *LicenseHandler {
	return &LicenseHandler{DB: db, Checker: checker, LCATopURL: lcaTopURL}
}

// remoteVerifyResult lca.top 远程验证结果
type remoteVerifyResult struct {
	Valid       bool   `json:"valid"`
	Error       string `json:"error"`
	LicenseType string `json:"license_type"`
	ExpiresAt   int64  `json:"expires_at"`
	IssuedAt    int64  `json:"issued_at"`
}

// verifyRemotely 调用 lca.top 远程验证授权码
func (h *LicenseHandler) verifyRemotely(licenseKey, machineID string) (*remoteVerifyResult, error) {
	url := fmt.Sprintf("%s/api/license/verify", h.LCATopURL)
	body, _ := json.Marshal(map[string]string{
		"license_key": licenseKey,
		"machine_id":  machineID,
	})
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("无法连接授权服务器 %s: %v", h.LCATopURL, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	var result remoteVerifyResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	return &result, nil
}

// Status 获取当前授权状态
func (h *LicenseHandler) Status(c *gin.Context) {
	machineID := license.GetMachineID()

	var lic model.License
	err := h.DB.Where("status = ?", 1).First(&lic).Error

	if err != nil {
		response.Success(c, gin.H{
			"activated":  false,
			"machine_id": machineID,
		})
		return
	}

	// 检查是否过期
	if lic.ExpiresAt != nil && time.Now().After(*lic.ExpiresAt) {
		h.DB.Model(&lic).Update("status", 2)
		response.Success(c, gin.H{
			"activated":    false,
			"machine_id":   machineID,
			"expired":      true,
			"license_type": lic.LicenseType,
			"expires_at":   lic.ExpiresAt,
		})
		return
	}

	c.Set("license_valid", true)
	response.Success(c, gin.H{
		"activated":    true,
		"machine_id":   machineID,
		"license_type": lic.LicenseType,
		"bound_at":     lic.BoundAt,
		"expires_at":   lic.ExpiresAt,
	})
}

// Activate 激活授权码
func (h *LicenseHandler) Activate(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"license_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ErrorCode, "参数错误")
		return
	}

	currentMachineID := license.GetMachineID()

	// ── 优先远程验证（lca.top 是权威数据源）──────────────────────
	if h.LCATopURL != "" {
		remoteResult, err := h.verifyRemotely(req.LicenseKey, currentMachineID)
		if err != nil {
			// 网络故障：降级本地 RSA 验签
			goto localVerify
		}
		if !remoteResult.Valid {
			msg := "授权码无效"
			if remoteResult.Error != "" {
				msg = remoteResult.Error
			}
			response.Error(c, response.ErrorCode, msg)
			return
		}
		// 远程验证通过，直接保存并返回
		h.saveAndRespond(c, req.LicenseKey, currentMachineID,
			remoteResult.LicenseType, remoteResult.ExpiresAt)
		return
	}

localVerify:
	// ── 本地 RSA 验签（降级或未配置远程地址）────────────────────
	if h.Checker == nil {
		response.Error(c, response.ErrorCode, "本地授权验签未配置（license.public_key），请联系管理员")
		return
	}
	result := h.Checker.Verify(req.LicenseKey)
	if !result.Valid {
		msg := "授权码无效"
		if result.IsExpired {
			msg = "授权码已过期"
		}
		response.Error(c, response.ErrorCode, msg)
		return
	}
	if result.Payload.MachineID != currentMachineID {
		response.Error(c, response.ErrorCode, "授权码与当前机器不匹配，请使用本机生成的授权码")
		return
	}
	h.saveAndRespond(c, req.LicenseKey, currentMachineID,
		result.Payload.Type, result.Payload.ExpiresAt)
}

// saveAndRespond 保存授权码记录并返回成功响应
func (h *LicenseHandler) saveAndRespond(c *gin.Context, licenseKey, machineID, licenseType string, expiresAtUnix int64) {
	// 检查是否已有有效授权
	var existing model.License
	if err := h.DB.Where("status = ?", 1).First(&existing).Error; err == nil {
		response.Error(c, response.ErrorCode, "已有有效授权码，请先解绑")
		return
	}

	now := time.Now()
	lic := model.License{
		LicenseKey:  licenseKey,
		MachineID:   machineID,
		LicenseType: licenseType,
		Status:      1,
		BoundAt:     &now,
	}
	if expiresAtUnix > 0 {
		expiresAt := time.Unix(expiresAtUnix, 0)
		lic.ExpiresAt = &expiresAt
	}

	if err := h.DB.Create(&lic).Error; err != nil {
		response.Error(c, response.ErrorCode, "授权码激活失败")
		return
	}

	c.Set("license_valid", true)
	response.Success(c, gin.H{
		"activated":    true,
		"license_type": lic.LicenseType,
		"expires_at":   lic.ExpiresAt,
	})
}

// Deactivate 解绑授权码
func (h *LicenseHandler) Deactivate(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	_ = userID

	var lic model.License
	if err := h.DB.Where("status = ?", 1).First(&lic).Error; err != nil {
		response.Error(c, response.ErrorCode, "没有已激活的授权码")
		return
	}

	h.DB.Model(&lic).Update("status", 0)
	c.Set("license_valid", false)
	response.Success(c, nil)
}
