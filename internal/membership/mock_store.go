package membership

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MockMembershipStore 内存实现的会员存储（用于测试和开发）
type MockMembershipStore struct {
	mu         sync.RWMutex
	membership map[string]*UserMembership
	usage      map[string]int // key: userID:date
	boostPacks map[string]*UserBoostPack
}

// NewMockStore 创建内存会员存储
func NewMockStore() *MockMembershipStore {
	return &MockMembershipStore{
		membership: make(map[string]*UserMembership),
		usage:      make(map[string]int),
		boostPacks: make(map[string]*UserBoostPack),
	}
}

// GetUserMembership 获取用户会员状态
func (s *MockMembershipStore) GetUserMembership(ctx context.Context, userID string) (*UserMembership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if m, ok := s.membership[userID]; ok {
		return m, nil
	}

	// 返回免费会员
	return &UserMembership{
		UserID:    userID,
		Tier:      TierFree,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// SaveUserMembership 保存用户会员状态
func (s *MockMembershipStore) SaveUserMembership(ctx context.Context, m *UserMembership) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m.UpdatedAt = time.Now()
	s.membership[m.UserID] = m
	return nil
}

// GetDailyUsage 获取用户当日使用量
func (s *MockMembershipStore) GetDailyUsage(ctx context.Context, userID, date string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := userID + ":" + date
	return s.usage[key], nil
}

// IncrDailyUsage 增加用户当日使用量
func (s *MockMembershipStore) IncrDailyUsage(ctx context.Context, userID, date string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userID + ":" + date
	s.usage[key]++
	return s.usage[key], nil
}

// GetUserBoostPack 获取用户加油包状态
func (s *MockMembershipStore) GetUserBoostPack(ctx context.Context, userID string) (*UserBoostPack, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if pack, ok := s.boostPacks[userID]; ok {
		return pack, nil
	}

	return &UserBoostPack{
		UserID:          userID,
		VideosRemaining: 0,
	}, nil
}

// SaveUserBoostPack 保存用户加油包状态
func (s *MockMembershipStore) SaveUserBoostPack(ctx context.Context, pack *UserBoostPack) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.boostPacks[pack.UserID] = pack
	return nil
}

// DecrBoostPack 消耗加油包配额
func (s *MockMembershipStore) DecrBoostPack(ctx context.Context, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pack, ok := s.boostPacks[userID]
	if !ok || !pack.IsValid() {
		return fmt.Errorf("boost pack invalid or expired")
	}

	pack.VideosRemaining--
	return nil
}

// SetUserTier 设置用户等级（测试辅助方法）
func (s *MockMembershipStore) SetUserTier(userID string, tier Tier) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.membership[userID] = &UserMembership{
		UserID:    userID,
		Tier:      tier,
		ExpiresAt: now.AddDate(1, 0, 0), // 1年后过期
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetUserTierWithExpiry 设置用户等级和过期时间（测试辅助方法）
func (s *MockMembershipStore) SetUserTierWithExpiry(userID string, tier Tier, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	s.membership[userID] = &UserMembership{
		UserID:    userID,
		Tier:      tier,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetDailyUsage 设置用户当日使用量（测试辅助方法）
func (s *MockMembershipStore) SetDailyUsage(userID, date string, count int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := userID + ":" + date
	s.usage[key] = count
}

// SetBoostPack 设置用户加油包（测试辅助方法）
func (s *MockMembershipStore) SetBoostPack(userID string, videos int, validDays int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.boostPacks[userID] = &UserBoostPack{
		UserID:          userID,
		VideosRemaining: videos,
		ExpiresAt:       time.Now().AddDate(0, 0, validDays),
		LastPurchaseAt:  time.Now(),
	}
}

// Reset 重置所有数据（测试辅助方法）
func (s *MockMembershipStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.membership = make(map[string]*UserMembership)
	s.usage = make(map[string]int)
	s.boostPacks = make(map[string]*UserBoostPack)
}
