package membership

import (
	"context"
	"time"
)

// MembershipStore 会员存储接口
type MembershipStore interface {
	// 会员状态
	GetUserMembership(ctx context.Context, userID string) (*UserMembership, error)
	SaveUserMembership(ctx context.Context, m *UserMembership) error

	// 每日配额
	GetDailyUsage(ctx context.Context, userID string, date string) (int, error)
	IncrDailyUsage(ctx context.Context, userID string, date string) (int, error)

	// 加油包
	GetUserBoostPack(ctx context.Context, userID string) (*UserBoostPack, error)
	SaveUserBoostPack(ctx context.Context, pack *UserBoostPack) error
	DecrBoostPack(ctx context.Context, userID string) error
}

// BoostPackType 加油包类型
type BoostPackType string

const (
	BoostPackSmall  BoostPackType = "small"
	BoostPackMedium BoostPackType = "medium"
	BoostPackLarge  BoostPackType = "large"
)

// BoostPackConfig 加油包配置
type BoostPackConfig struct {
	Type        BoostPackType `json:"type"`
	Name        string        `json:"name"`
	Price       float64       `json:"price"`
	Videos      int           `json:"videos"`
	ValidDays   int           `json:"valid_days"`
	Description string        `json:"description"`
}

// DefaultBoostPackConfigs 默认加油包配置
var DefaultBoostPackConfigs = map[BoostPackType]BoostPackConfig{
	BoostPackSmall: {
		Type:        BoostPackSmall,
		Name:        "小加油包",
		Price:       6.9,
		Videos:      10,
		ValidDays:   7,
		Description: "10个视频配额，7天有效",
	},
	BoostPackMedium: {
		Type:        BoostPackMedium,
		Name:        "中加油包",
		Price:       19.9,
		Videos:      30,
		ValidDays:   15,
		Description: "30个视频配额，15天有效",
	},
	BoostPackLarge: {
		Type:        BoostPackLarge,
		Name:        "大加油包",
		Price:       39.9,
		Videos:      80,
		ValidDays:   30,
		Description: "80个视频配额，30天有效",
	},
}

// UserBoostPack 用户加油包状态
type UserBoostPack struct {
	UserID          string    `json:"user_id"`
	VideosRemaining int       `json:"videos_remaining"`
	ExpiresAt       time.Time `json:"expires_at"`
	LastPurchaseAt  time.Time `json:"last_purchase_at"`
}

// IsValid 检查加油包是否有效
func (b *UserBoostPack) IsValid() bool {
	return b.VideosRemaining > 0 && time.Now().Before(b.ExpiresAt)
}

// DaysUntilExpiry 返回距离过期的天数
func (b *UserBoostPack) DaysUntilExpiry() int {
	if !b.IsValid() {
		return 0
	}
	return int(time.Until(b.ExpiresAt).Hours() / 24)
}

// GetBoostPackConfig 获取指定类型的加油包配置
func GetBoostPackConfig(packType BoostPackType) (BoostPackConfig, bool) {
	config, ok := DefaultBoostPackConfigs[packType]
	return config, ok
}

// GetAllBoostPackConfigs 获取所有加油包配置
func GetAllBoostPackConfigs() []BoostPackConfig {
	configs := make([]BoostPackConfig, 0, len(DefaultBoostPackConfigs))
	for _, config := range DefaultBoostPackConfigs {
		configs = append(configs, config)
	}
	return configs
}
