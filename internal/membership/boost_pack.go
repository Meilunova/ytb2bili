package membership

import (
	"context"
	"fmt"
	"time"
)

// BoostPackService 加油包服务
type BoostPackService struct {
	store MembershipStore
}

// NewBoostPackService 创建加油包服务
func NewBoostPackService(store MembershipStore) *BoostPackService {
	return &BoostPackService{store: store}
}

// BoostPackStatus 加油包状态
type BoostPackStatus struct {
	HasPack         bool      `json:"has_pack"`
	VideosRemaining int       `json:"videos_remaining"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	DaysRemaining   int       `json:"days_remaining"`
}

// GetStatus 获取用户加油包状态
func (s *BoostPackService) GetStatus(ctx context.Context, userID string) (*BoostPackStatus, error) {
	pack, err := s.store.GetUserBoostPack(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取加油包失败: %w", err)
	}

	if pack == nil || !pack.IsValid() {
		return &BoostPackStatus{
			HasPack:         false,
			VideosRemaining: 0,
			DaysRemaining:   0,
		}, nil
	}

	return &BoostPackStatus{
		HasPack:         true,
		VideosRemaining: pack.VideosRemaining,
		ExpiresAt:       pack.ExpiresAt,
		DaysRemaining:   pack.DaysUntilExpiry(),
	}, nil
}

// PurchaseResult 购买结果
type PurchaseResult struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	VideosAdded int       `json:"videos_added"`
	TotalVideos int       `json:"total_videos"`
	ExpiresAt   time.Time `json:"expires_at"`
	NewPurchase bool      `json:"new_purchase"` // 是否为新购买（非叠加）
}

// Purchase 购买加油包
func (s *BoostPackService) Purchase(ctx context.Context, userID string, packType BoostPackType) (*PurchaseResult, error) {
	// 获取加油包配置
	config, ok := GetBoostPackConfig(packType)
	if !ok {
		return nil, fmt.Errorf("无效的加油包类型: %s", packType)
	}

	// 获取当前加油包状态
	currentPack, err := s.store.GetUserBoostPack(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("获取当前加油包失败: %w", err)
	}

	now := time.Now()
	var newPack *UserBoostPack
	var newPurchase bool

	if currentPack != nil && currentPack.IsValid() {
		// 叠加到现有加油包
		newPack = &UserBoostPack{
			UserID:          userID,
			VideosRemaining: currentPack.VideosRemaining + config.Videos,
			ExpiresAt:       currentPack.ExpiresAt, // 保持原过期时间
			LastPurchaseAt:  now,
		}
		// 如果新包有效期更长，延长过期时间
		newExpiry := now.AddDate(0, 0, config.ValidDays)
		if newExpiry.After(currentPack.ExpiresAt) {
			newPack.ExpiresAt = newExpiry
		}
		newPurchase = false
	} else {
		// 新购买
		newPack = &UserBoostPack{
			UserID:          userID,
			VideosRemaining: config.Videos,
			ExpiresAt:       now.AddDate(0, 0, config.ValidDays),
			LastPurchaseAt:  now,
		}
		newPurchase = true
	}

	// 保存加油包
	if err := s.store.SaveUserBoostPack(ctx, newPack); err != nil {
		return nil, fmt.Errorf("保存加油包失败: %w", err)
	}

	return &PurchaseResult{
		Success:     true,
		Message:     fmt.Sprintf("成功购买%s", config.Name),
		VideosAdded: config.Videos,
		TotalVideos: newPack.VideosRemaining,
		ExpiresAt:   newPack.ExpiresAt,
		NewPurchase: newPurchase,
	}, nil
}

// Consume 消耗加油包配额
func (s *BoostPackService) Consume(ctx context.Context, userID string) error {
	return s.store.DecrBoostPack(ctx, userID)
}

// HasValidPack 检查用户是否有有效加油包
func (s *BoostPackService) HasValidPack(ctx context.Context, userID string) bool {
	pack, err := s.store.GetUserBoostPack(ctx, userID)
	if err != nil {
		return false
	}
	return pack != nil && pack.IsValid()
}

// GetAvailableVideos 获取加油包可用视频数
func (s *BoostPackService) GetAvailableVideos(ctx context.Context, userID string) int {
	pack, err := s.store.GetUserBoostPack(ctx, userID)
	if err != nil || pack == nil || !pack.IsValid() {
		return 0
	}
	return pack.VideosRemaining
}
