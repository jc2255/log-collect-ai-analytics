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
// 重要：必须与 LicenseCheck 中间件判断逻辑一致（RSA 验签 + 机器ID对比），
// 否则会出现“status 说已激活，其他接口返回 40301”的不一致，导致前端不弹授权框。
func (h *LicenseHandler) Status(c *gin.Context) {
	machineID := license.GetMachineID()

	var lic model.License
	// 取最新一条记录
	err := h.DB.Order("id DESC").First(&lic).Error
	if err != nil {
		response.Success(c, gin.H{
			"activated":  false,
			"machine_id": machineID,
		})
		return
	}

	// status=0 表示用户已主动解绑，直接返回未激活
	if lic.Status == 0 {
		response.Success(c, gin.H{
			"activated":  false,
			"machine_id": machineID,
		})
		return
	}

	// ── 优先远程验证（lca.top 权威数据源）─────────────────────
	// 适用场景：本地 public_key 与 lca.top 签发用私钥不配对（典型）
	// 使 Status 接口与 Activate / LicenseCheck 中间件保持一致验证逻辑
	if h.LCATopURL != "" {
		remoteResult, rerr := h.verifyRemotely(lic.LicenseKey, machineID)
		if rerr == nil {
			if remoteResult.Valid {
				c.Set("license_valid", true)
				response.Success(c, gin.H{
					"activated":    true,
					"machine_id":   machineID,
					"license_type": lic.LicenseType,
					"bound_at":     lic.BoundAt,
					"expires_at":   lic.ExpiresAt,
				})
				return
			}
			// 远程明确返回 invalid
			response.Success(c, gin.H{
				"activated":    false,
				"machine_id":   machineID,
				"reason":       "remote: " + remoteResult.Error,
				"license_type": lic.LicenseType,
				"expires_at":   lic.ExpiresAt,
			})
			return
		}
		// 远程网络故障 → 降级本地 RSA 验签
	}

	// 未配置验签器（本地验签降级不可用）时，仅依赖 status 字段作为兑底
	if h.Checker == nil {
		activated := lic.Status == 1
		result := gin.H{
			"activated":    activated,
			"machine_id":   machineID,
			"license_type": lic.LicenseType,
			"bound_at":     lic.BoundAt,
			"expires_at":   lic.ExpiresAt,
		}
		if !activated {
			result["reason"] = "verifier disabled"
		}
		response.Success(c, result)
		return
	}

	// RSA 验签 + 过期检查
	verifyResult := h.Checker.Verify(lic.LicenseKey)
	if !verifyResult.Valid {
		reason := "signature invalid"
		if verifyResult.IsExpired {
			reason = "expired"
		}
		response.Success(c, gin.H{
			"activated":    false,
			"machine_id":   machineID,
			"expired":      verifyResult.IsExpired,
			"reason":       reason,
			"license_type": lic.LicenseType,
			"expires_at":   lic.ExpiresAt,
		})
		return
	}

	// 机器ID不匹配（客户迁移机器或容器重建后丢失机器ID文件后出现）
	if verifyResult.Payload != nil && verifyResult.Payload.MachineID != machineID {
		response.Success(c, gin.H{
			"activated":     false,
			"machine_id":    machineID,
			"reason":        "machine id mismatch",
			"bound_machine": verifyResult.Payload.MachineID,
			"license_type":  lic.LicenseType,
			"expires_at":    lic.ExpiresAt,
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
	// 检查是否已有「真正有效」的授权（status=1 且未过期）
	// 过期授权不再阻塞激活，直接覆盖
	var existing model.License
	if err := h.DB.Where("status = ?", 1).First(&existing).Error; err == nil {
		if existing.ExpiresAt == nil || existing.ExpiresAt.After(time.Now()) {
			// 永久授权（ExpiresAt==nil）或尚未过期 → 拒绝重复激活
			response.Error(c, response.ErrorCode, "已有有效授权码，请先解绑")
			return
		}
		// 已过期 → 将旧记录标记为过期，允许激活新的
		h.DB.Model(&existing).Update("status", 2)
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

	// 激活成功后立即失效中间件缓存，下个业务请求会重新验签
	middleware.InvalidateLicenseCache()

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
	middleware.InvalidateLicenseCache()
	c.Set("license_valid", false)
	response.Success(c, nil)
}
