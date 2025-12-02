package membership

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefixMembership = "ytb2bili:membership:"
	keyPrefixUsage      = "ytb2bili:usage:"
	keyPrefixBoostPack  = "ytb2bili:boost:"
	membershipCacheTTL  = 5 * time.Minute
	usageKeyTTL         = 48 * time.Hour // 保留2天，确保跨日查询
)

// RedisMembershipStore Redis 实现的会员存储
type RedisMembershipStore struct {
	client *redis.Client
}

// NewRedisMembershipStore 创建 Redis 会员存储
func NewRedisMembershipStore(client *redis.Client) *RedisMembershipStore {
	return &RedisMembershipStore{client: client}
}

// GetUserMembership 获取用户会员状态
func (s *RedisMembershipStore) GetUserMembership(ctx context.Context, userID string) (*UserMembership, error) {
	key := keyPrefixMembership + userID
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// 用户不存在，返回免费会员
		return &UserMembership{
			UserID:    userID,
			Tier:      TierFree,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get membership failed: %w", err)
	}

	var m UserMembership
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal membership failed: %w", err)
	}
	return &m, nil
}

// SaveUserMembership 保存用户会员状态
func (s *RedisMembershipStore) SaveUserMembership(ctx context.Context, m *UserMembership) error {
	key := keyPrefixMembership + m.UserID
	m.UpdatedAt = time.Now()

	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal membership failed: %w", err)
	}

	// 计算 TTL：如果有过期时间，使用过期时间+1小时；否则使用默认缓存时间
	ttl := membershipCacheTTL
	if !m.ExpiresAt.IsZero() && m.ExpiresAt.After(time.Now()) {
		ttl = time.Until(m.ExpiresAt) + time.Hour
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

// GetDailyUsage 获取用户当日使用量
func (s *RedisMembershipStore) GetDailyUsage(ctx context.Context, userID, date string) (int, error) {
	key := fmt.Sprintf("%s%s:%s", keyPrefixUsage, userID, date)
	count, err := s.client.Get(ctx, key).Int()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("redis get usage failed: %w", err)
	}
	return count, nil
}

// IncrDailyUsage 增加用户当日使用量
func (s *RedisMembershipStore) IncrDailyUsage(ctx context.Context, userID, date string) (int, error) {
	key := fmt.Sprintf("%s%s:%s", keyPrefixUsage, userID, date)

	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, usageKeyTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("redis incr usage failed: %w", err)
	}

	return int(incr.Val()), nil
}

// GetUserBoostPack 获取用户加油包状态
func (s *RedisMembershipStore) GetUserBoostPack(ctx context.Context, userID string) (*UserBoostPack, error) {
	key := keyPrefixBoostPack + userID
	data, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		// 用户没有加油包
		return &UserBoostPack{
			UserID:          userID,
			VideosRemaining: 0,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get boost pack failed: %w", err)
	}

	var pack UserBoostPack
	if err := json.Unmarshal(data, &pack); err != nil {
		return nil, fmt.Errorf("unmarshal boost pack failed: %w", err)
	}
	return &pack, nil
}

// SaveUserBoostPack 保存用户加油包状态
func (s *RedisMembershipStore) SaveUserBoostPack(ctx context.Context, pack *UserBoostPack) error {
	key := keyPrefixBoostPack + pack.UserID

	data, err := json.Marshal(pack)
	if err != nil {
		return fmt.Errorf("marshal boost pack failed: %w", err)
	}

	// TTL 设置为过期时间+1小时，最少1小时
	ttl := time.Until(pack.ExpiresAt) + time.Hour
	if ttl < time.Hour {
		ttl = time.Hour
	}

	return s.client.Set(ctx, key, data, ttl).Err()
}

// DecrBoostPack 消耗加油包配额
func (s *RedisMembershipStore) DecrBoostPack(ctx context.Context, userID string) error {
	pack, err := s.GetUserBoostPack(ctx, userID)
	if err != nil {
		return err
	}

	if !pack.IsValid() {
		return fmt.Errorf("boost pack invalid or expired")
	}

	pack.VideosRemaining--
	return s.SaveUserBoostPack(ctx, pack)
}

// Ping 检查 Redis 连接
func (s *RedisMembershipStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

// Close 关闭 Redis 连接
func (s *RedisMembershipStore) Close() error {
	return s.client.Close()
}
