package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidAppID      = errors.New("无效的 App ID")
	ErrInvalidAppSecret  = errors.New("无效的 App Secret")
	ErrAppDisabled       = errors.New("应用已禁用")
	ErrIPNotAllowed      = errors.New("IP 不在白名单中")
	ErrRateLimitExceeded = errors.New("请求频率超限")
)

// GenerateAppID 生成 App ID
func GenerateAppID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("app_%s", hex.EncodeToString(bytes))
}

// GenerateAppSecret 生成 App Secret
func GenerateAppSecret() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GenerateSignature 生成请求签名
// 签名算法: HMAC-SHA256(appId + timestamp + nonce, appSecret)
func GenerateSignature(appID, appSecret, timestamp, nonce string) string {
	message := appID + timestamp + nonce
	h := hmac.New(sha256.New, []byte(appSecret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature 验证请求签名
func VerifySignature(appID, appSecret, timestamp, nonce, signature string) bool {
	expected := GenerateSignature(appID, appSecret, timestamp, nonce)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ValidateTimestamp 验证时间戳 (防重放攻击)
func ValidateTimestamp(timestamp string, maxAge time.Duration) bool {
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		// 尝试 Unix 时间戳
		var unix int64
		_, err = fmt.Sscanf(timestamp, "%d", &unix)
		if err != nil {
			return false
		}
		ts = time.Unix(unix, 0)
	}

	diff := time.Since(ts)
	if diff < 0 {
		diff = -diff
	}
	return diff <= maxAge
}

// AppCredentials App 认证凭据
type AppCredentials struct {
	AppID     string `json:"app_id"`
	Timestamp string `json:"timestamp"`
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

// SimpleAppAuth 简单 App 认证 (仅验证 appId + appSecret)
// 适用于服务端到服务端的调用
type SimpleAppAuth struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}
