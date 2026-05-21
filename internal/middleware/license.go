package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cj/log-collect-ai-analytics/internal/model"
	"github.com/cj/log-collect-ai-analytics/internal/pkg/license"
)

// 授权检查白名单路径(无需授权即可访问)
var licenseWhitelist = []string{
	"/api/v1/auth/login",
	"/api/v1/captcha",
	"/api/v1/license/",
	"/api/v1/auth/userinfo",
	"/health",
}

// licenseDB 授权检查使用的数据库实例
var licenseDB *gorm.DB

// licenseVerifier RSA验签器（用于运行时重新验签，防止DB字段被篡改）
var licenseVerifier *license.Verifier

// licenseLCATopURL lca.top 远程验证地址（权威数据源）
// 配置后中间件会优先调远程验证，本地RSA仅作降级兜底
var licenseLCATopURL string

// 验签结果缓存（5分钟刷新，避免每个请求都做 RSA 验签）
var (
	licenseCache     licenseCacheEntry
	licenseCacheLock sync.RWMutex
)

type licenseCacheEntry struct {
	valid   bool
	checkAt time.Time
	reason  string // 失败原因（用于日志/调试）
}

const licenseCacheTTL = 1 * time.Minute

// InitLicenseDB 初始化授权检查数据库
func InitLicenseDB(db *gorm.DB) {
	licenseDB = db
}

// InitLicenseVerifier 注入 RSA 验签器，让中间件具备从密文重新验签的能力
// 强烈建议在生产环境配置，否则会退化为只查 DB status 字段（不安全）
func InitLicenseVerifier(v *license.Verifier) {
	licenseVerifier = v
}

// InitLicenseLCATopURL 注入 lca.top 远程验证地址
// 当本地公钥与 lca.top 签发用私钥不配对时（典型场景），
// 中间件可走远程验证避免本地 RSA 失败导致死循环（激活后刷新又要激活）
func InitLicenseLCATopURL(url string) {
	licenseLCATopURL = url
}

// remoteVerify 调用 lca.top 远程验证授权码（权威）
type remoteVerifyResp struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}

func verifyLicenseRemotely(licenseKey, machineID string) (bool, string) {
	if licenseLCATopURL == "" {
		return false, "remote url not configured"
	}
	url := fmt.Sprintf("%s/api/license/verify", licenseLCATopURL)
	body, _ := json.Marshal(map[string]string{
		"license_key": licenseKey,
		"machine_id":  machineID,
	})
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, "remote unreachable: " + err.Error()
	}
	defer resp.Body.Close()
	var r remoteVerifyResp
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return false, "remote bad response"
	}
	if !r.Valid {
		return false, "remote: " + r.Error
	}
	return true, ""
}

// validateActiveLicense 从 DB 取出已激活的 license_key 重新做 RSA 验签
// 不信任 DB 中的 status / expires_at 字段（客户可手改 MySQL 绕过）
// 一切以签名保护的 payload 为准
func validateActiveLicense() (bool, string) {
	if licenseDB == nil {
		return false, "license db not initialized"
	}

	var lic model.License
	// 注意：故意不加 status=1 条件，取最新的一条 license_key 实际验签
	// 这样客户即使把 status 改成 0/2 也不影响（验签为准）
	if err := licenseDB.Order("id DESC").First(&lic).Error; err != nil {
		return false, "no license record"
	}

	currentMachineID := license.GetMachineID()

	// ── 优先远程验证（lca.top 权威数据源）─────────────────────
	// 适用场景：客户本地公钥与 lca.top 签发用私钥不配对（典型）
	if licenseLCATopURL != "" {
		ok, reason := verifyLicenseRemotely(lic.LicenseKey, currentMachineID)
		if ok {
			return true, ""
		}
		// 远程明确返回 invalid（非网络故障）→ 直接拒绝
		if !strings.HasPrefix(reason, "remote unreachable") && !strings.HasPrefix(reason, "remote bad response") {
			return false, reason
		}
		// 网络故障 → 降级本地 RSA 验签
	}

	// ── 本地 RSA 验签（远程未配置或网络故障）──────────────────
	// 没有验签器 → 退化为旧行为（仅查 DB status）
	if licenseVerifier == nil {
		if lic.Status == 1 {
			return true, ""
		}
		return false, "no active license (verifier disabled)"
	}

	// RSA 验签 license_key（核心防御：客户没有私钥，无法伪造签名）
	result := licenseVerifier.Verify(lic.LicenseKey)
	if !result.Valid {
		if result.IsExpired {
			return false, "license expired"
		}
		if result.Err != nil {
			return false, "signature invalid: " + result.Err.Error()
		}
		return false, "license invalid"
	}

	// 比对机器指纹（payload 里的 machine_id 是签名保护的真值）
	if result.Payload != nil {
		if result.Payload.MachineID != currentMachineID {
			return false, "machine id mismatch"
		}
	}

	return true, ""
}

// LicenseCheck 授权码检查中间件（每个非白名单请求都会触发）
// 安全策略：DB 中的 status / expires_at 不可信，每次（带缓存）从 license_key 重新 RSA 验签
func LicenseCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// 白名单放行
		for _, prefix := range licenseWhitelist {
			if strings.HasPrefix(path, prefix) || path == prefix {
				c.Next()
				return
			}
		}

		// 读缓存（避免每次请求都做 RSA 验签，性能考虑）
		licenseCacheLock.RLock()
		cached := licenseCache
		licenseCacheLock.RUnlock()

		now := time.Now()
		if !cached.checkAt.IsZero() && now.Sub(cached.checkAt) < licenseCacheTTL {
			if cached.valid {
				c.Next()
				return
			}
			c.JSON(http.StatusForbidden, gin.H{
				"code":    40301,
				"message": "license_required",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 缓存过期或首次：重新验签
		valid, reason := validateActiveLicense()

		licenseCacheLock.Lock()
		licenseCache = licenseCacheEntry{
			valid:   valid,
			checkAt: now,
			reason:  reason,
		}
		licenseCacheLock.Unlock()

		if valid {
			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    40301,
			"message": "license_required",
			"data":    nil,
		})
		c.Abort()
	}
}

// InvalidateLicenseCache 在 Activate / Deactivate 后调用，立刻失效缓存让下一个请求重新验签
func InvalidateLicenseCache() {
	licenseCacheLock.Lock()
	licenseCache = licenseCacheEntry{}
	licenseCacheLock.Unlock()
}
