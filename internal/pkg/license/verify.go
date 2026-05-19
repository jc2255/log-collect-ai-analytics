package license

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"time"
)

// LicensePayload 授权码payload结构
type LicensePayload struct {
	Product   string `json:"product"`    // 产品标识: "lca"
	MachineID string `json:"machine_id"` // 绑定的机器指纹
	Type      string `json:"type"`       // monthly/yearly/permanent
	IssuedAt  int64  `json:"issued_at"`  // 签发时间(unix)
	ExpiresAt int64  `json:"expires_at"` // 到期时间(unix), permanent为0
}

// VerifyResult 验证结果
type VerifyResult struct {
	Valid     bool
	Payload   *LicensePayload
	Err       error
	IsExpired bool
	MachineID string // 授权码中绑定的machine_id
}

// Verifier RSA公钥验签器
type Verifier struct {
	publicKey *rsa.PublicKey
}

// NewVerifier 从PEM格式公钥创建验签器
func NewVerifier(publicKeyPEM string) (*Verifier, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return &Verifier{publicKey: rsaPub}, nil
}

// Verify 验证授权码
// licenseKey格式: base64(payload).base64(signature)
func (v *Verifier) Verify(licenseKey string) *VerifyResult {
	result := &VerifyResult{}

	// 拆分payload和signature
	parts := splitLicenseKey(licenseKey)
	if len(parts) != 2 {
		result.Err = errors.New("invalid license key format")
		return result
	}

	payloadB64 := parts[0]
	sigB64 := parts[1]

	// 解码payload
	payloadBytes, err := base64.StdEncoding.DecodeString(payloadB64)
	if err != nil {
		result.Err = errors.New("failed to decode payload: " + err.Error())
		return result
	}

	var payload LicensePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		result.Err = errors.New("failed to parse payload: " + err.Error())
		return result
	}
	result.Payload = &payload
	result.MachineID = payload.MachineID

	// 解码签名
	sigBytes, err := base64.StdEncoding.DecodeString(sigB64)
	if err != nil {
		result.Err = errors.New("failed to decode signature: " + err.Error())
		return result
	}

	// 验证RSA签名
	hashed := sha256.Sum256(payloadBytes)
	err = rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hashed[:], sigBytes)
	if err != nil {
		result.Err = errors.New("signature verification failed: " + err.Error())
		return result
	}

	// 检查产品标识
	if payload.Product != "lca" {
		result.Err = errors.New("invalid product identifier")
		return result
	}

	// 检查过期时间(permanent类型expires_at为0，永不过期)
	if payload.ExpiresAt > 0 {
		expiresAt := time.Unix(payload.ExpiresAt, 0)
		if time.Now().After(expiresAt) {
			result.IsExpired = true
			result.Err = errors.New("license has expired")
			return result
		}
	}

	result.Valid = true
	return result
}

// splitLicenseKey 拆分授权码
func splitLicenseKey(key string) []string {
	// 格式: payload_base64.signature_base64
	// 找到最后一个点号来分割(因为base64本身不含点号)
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '.' {
			return []string{key[:i], key[i+1:]}
		}
	}
	return nil
}
