package membership

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HybridMembershipStore 混合存储（Redis缓存 + 数据库持久化）
// 会员信息：数据库为主，Redis缓存
// 每日使用量：Redis为主（高频读写）
// 加油包：数据库为主，Redis缓存
type HybridMembershipStore struct {
	db    *DBMembershipStore
	redis *RedisMembershipStore
}

// NewHybridMembershipStore 创建混合存储
func NewHybridMembershipStore(db *gorm.DB, redisClient *redis.Client) *HybridMembershipStore {
	return &HybridMembershipStore{
		db:    NewDBMembershipStore(db),
		redis: NewRedisMembershipStore(redisClient),
	}
}

// GetUserMembership 获取用户会员状态（优先从缓存读取）
func (s *HybridMembershipStore) GetUserMembership(ctx context.Context, userID string) (*UserMembership, error) {
	// 先尝试从 Redis 缓存读取
	m, err := s.redis.GetUserMembership(ctx, userID)
	if err == nil && m.Tier != TierFree {
		// 缓存命中且不是默认值
		return m, nil
	}

	// 缓存未命中，从数据库读取
	m, err = s.db.GetUserMembership(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存（异步，不阻塞）
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.redis.SaveUserMembership(ctx, m)
	}()

	return m, nil
}

// SaveUserMembership 保存用户会员状态（同时写入数据库和缓存）
func (s *HybridMembershipStore) SaveUserMembership(ctx context.Context, m *UserMembership) error {
	// 先写数据库
	if err := s.db.SaveUserMembership(ctx, m); err != nil {
		return err
	}

	// 再更新缓存
	return s.redis.SaveUserMembership(ctx, m)
}

// GetDailyUsage 获取用户当日使用量（从 Redis 读取，高频操作）
func (s *HybridMembershipStore) GetDailyUsage(ctx context.Context, userID string, date string) (int, error) {
	// 优先从 Redis 读取
	count, err := s.redis.GetDailyUsage(ctx, userID, date)
	if err == nil && count > 0 {
		return count, nil
	}

	// Redis 没有，从数据库读取
	count, err = s.db.GetDailyUsage(ctx, userID, date)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// IncrDailyUsage 增加用户当日使用量（Redis 为主，定期同步到数据库）
func (s *HybridMembershipStore) IncrDailyUsage(ctx context.Context, userID string, date string) (int, error) {
	// 在 Redis 中增加
	count, err := s.redis.IncrDailyUsage(ctx, userID, date)
	if err != nil {
		// Redis 失败，降级到数据库
		return s.db.IncrDailyUsage(ctx, userID, date)
	}

	// 异步同步到数据库（每10次同步一次，减少数据库压力）
	if count%10 == 0 {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.db.IncrDailyUsage(ctx, userID, date)
		}()
	}

	return count, nil
}

// GetUserBoostPack 获取用户加油包状态
func (s *HybridMembershipStore) GetUserBoostPack(ctx context.Context, userID string) (*UserBoostPack, error) {
	// 先尝试从 Redis 缓存读取
	pack, err := s.redis.GetUserBoostPack(ctx, userID)
	if err == nil && pack.VideosRemaining > 0 {
		return pack, nil
	}

	// 从数据库读取
	pack, err = s.db.GetUserBoostPack(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if pack.IsValid() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			s.redis.SaveUserBoostPack(ctx, pack)
		}()
	}

	return pack, nil
}

// SaveUserBoostPack 保存用户加油包状态
func (s *HybridMembershipStore) SaveUserBoostPack(ctx context.Context, pack *UserBoostPack) error {
	// 先写数据库
	if err := s.db.SaveUserBoostPack(ctx, pack); err != nil {
		return err
	}

	// 再更新缓存
	return s.redis.SaveUserBoostPack(ctx, pack)
}

// DecrBoostPack 消耗加油包配额
func (s *HybridMembershipStore) DecrBoostPack(ctx context.Context, userID string) error {
	// 先在数据库中原子减少
	if err := s.db.DecrBoostPack(ctx, userID); err != nil {
		return err
	}

	// 更新缓存
	return s.redis.DecrBoostPack(ctx, userID)
}

// AutoMigrate 自动迁移数据库
func (s *HybridMembershipStore) AutoMigrate() error {
	return s.db.AutoMigrate()
}

// MigrateExistingUsers 迁移现有用户
func (s *HybridMembershipStore) MigrateExistingUsers(ctx context.Context) error {
	return s.db.MigrateExistingUsers(ctx)
}

// SyncDailyUsageToDB 同步每日使用量到数据库（可由定时任务调用）
func (s *HybridMembershipStore) SyncDailyUsageToDB(ctx context.Context, userID string, date string) error {
	count, err := s.redis.GetDailyUsage(ctx, userID, date)
	if err != nil {
		return err
	}

	// 直接更新数据库
	return s.db.db.WithContext(ctx).
		Model(&struct{ DailyUsageCount int }{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"daily_usage_count": count,
			"daily_usage_date":  date,
		}).Error
}
