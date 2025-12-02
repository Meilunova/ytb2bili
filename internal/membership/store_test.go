package membership

import (
	"context"
	"testing"
	"time"
)

func TestMockStore_GetUserMembership(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// 新用户应该返回免费会员
	m, err := store.GetUserMembership(ctx, "new-user")
	if err != nil {
		t.Fatalf("GetUserMembership failed: %v", err)
	}
	if m.Tier != TierFree {
		t.Errorf("New user should be free tier, got %s", m.Tier)
	}

	// 设置用户等级
	store.SetUserTier("test-user", TierPro)
	m, err = store.GetUserMembership(ctx, "test-user")
	if err != nil {
		t.Fatalf("GetUserMembership failed: %v", err)
	}
	if m.Tier != TierPro {
		t.Errorf("Expected Pro tier, got %s", m.Tier)
	}
}

func TestMockStore_SaveUserMembership(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	membership := &UserMembership{
		UserID:    "test-user",
		Tier:      TierBasic,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	err := store.SaveUserMembership(ctx, membership)
	if err != nil {
		t.Fatalf("SaveUserMembership failed: %v", err)
	}

	// 验证保存成功
	m, _ := store.GetUserMembership(ctx, "test-user")
	if m.Tier != TierBasic {
		t.Errorf("Expected Basic tier, got %s", m.Tier)
	}
}

func TestMockStore_DailyUsage(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()
	userID := "test-user"
	date := time.Now().Format("2006-01-02")

	// 初始使用量为 0
	usage, err := store.GetDailyUsage(ctx, userID, date)
	if err != nil {
		t.Fatalf("GetDailyUsage failed: %v", err)
	}
	if usage != 0 {
		t.Errorf("Initial usage should be 0, got %d", usage)
	}

	// 增加使用量
	newUsage, err := store.IncrDailyUsage(ctx, userID, date)
	if err != nil {
		t.Fatalf("IncrDailyUsage failed: %v", err)
	}
	if newUsage != 1 {
		t.Errorf("Usage after incr should be 1, got %d", newUsage)
	}

	// 再次增加
	newUsage, _ = store.IncrDailyUsage(ctx, userID, date)
	if newUsage != 2 {
		t.Errorf("Usage after second incr should be 2, got %d", newUsage)
	}

	// 验证获取
	usage, _ = store.GetDailyUsage(ctx, userID, date)
	if usage != 2 {
		t.Errorf("Expected usage 2, got %d", usage)
	}
}

func TestMockStore_BoostPack(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()
	userID := "test-user"

	// 初始没有加油包
	pack, err := store.GetUserBoostPack(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserBoostPack failed: %v", err)
	}
	if pack.VideosRemaining != 0 {
		t.Errorf("Initial boost pack should have 0 videos, got %d", pack.VideosRemaining)
	}
	if pack.IsValid() {
		t.Error("Empty boost pack should not be valid")
	}

	// 设置加油包
	store.SetBoostPack(userID, 10, 7)
	pack, _ = store.GetUserBoostPack(ctx, userID)
	if pack.VideosRemaining != 10 {
		t.Errorf("Expected 10 videos, got %d", pack.VideosRemaining)
	}
	if !pack.IsValid() {
		t.Error("Boost pack should be valid")
	}

	// 消耗加油包
	err = store.DecrBoostPack(ctx, userID)
	if err != nil {
		t.Fatalf("DecrBoostPack failed: %v", err)
	}
	pack, _ = store.GetUserBoostPack(ctx, userID)
	if pack.VideosRemaining != 9 {
		t.Errorf("Expected 9 videos after decr, got %d", pack.VideosRemaining)
	}
}

func TestMockStore_DecrBoostPack_Invalid(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// 没有加油包时消耗应该失败
	err := store.DecrBoostPack(ctx, "no-pack-user")
	if err == nil {
		t.Error("DecrBoostPack should fail when no pack exists")
	}

	// 过期的加油包消耗应该失败
	store.boostPacks["expired-user"] = &UserBoostPack{
		UserID:          "expired-user",
		VideosRemaining: 5,
		ExpiresAt:       time.Now().Add(-24 * time.Hour), // 已过期
	}
	err = store.DecrBoostPack(ctx, "expired-user")
	if err == nil {
		t.Error("DecrBoostPack should fail when pack is expired")
	}
}

func TestMockStore_SetUserTierWithExpiry(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	expiry := time.Now().Add(7 * 24 * time.Hour)
	store.SetUserTierWithExpiry("test-user", TierPro, expiry)

	m, _ := store.GetUserMembership(ctx, "test-user")
	if m.Tier != TierPro {
		t.Errorf("Expected Pro tier, got %s", m.Tier)
	}
	if !m.ExpiresAt.Equal(expiry) {
		t.Errorf("Expiry time mismatch")
	}
}

func TestMockStore_Reset(t *testing.T) {
	store := NewMockStore()
	ctx := context.Background()

	// 添加一些数据
	store.SetUserTier("user1", TierPro)
	store.SetBoostPack("user1", 10, 7)
	store.IncrDailyUsage(ctx, "user1", "2024-01-01")

	// 重置
	store.Reset()

	// 验证数据已清空
	m, _ := store.GetUserMembership(ctx, "user1")
	if m.Tier != TierFree {
		t.Error("After reset, user should be free tier")
	}

	pack, _ := store.GetUserBoostPack(ctx, "user1")
	if pack.VideosRemaining != 0 {
		t.Error("After reset, boost pack should be empty")
	}

	usage, _ := store.GetDailyUsage(ctx, "user1", "2024-01-01")
	if usage != 0 {
		t.Error("After reset, usage should be 0")
	}
}

func TestBoostPackIsValid(t *testing.T) {
	tests := []struct {
		name   string
		pack   *UserBoostPack
		expect bool
	}{
		{
			name:   "empty pack",
			pack:   &UserBoostPack{VideosRemaining: 0, ExpiresAt: time.Now().Add(time.Hour)},
			expect: false,
		},
		{
			name:   "expired pack",
			pack:   &UserBoostPack{VideosRemaining: 10, ExpiresAt: time.Now().Add(-time.Hour)},
			expect: false,
		},
		{
			name:   "valid pack",
			pack:   &UserBoostPack{VideosRemaining: 10, ExpiresAt: time.Now().Add(time.Hour)},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pack.IsValid(); got != tt.expect {
				t.Errorf("IsValid() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestBoostPackConfig(t *testing.T) {
	// 测试获取加油包配置
	config, ok := GetBoostPackConfig(BoostPackSmall)
	if !ok {
		t.Fatal("Small boost pack config should exist")
	}
	if config.Videos != 10 {
		t.Errorf("Small pack should have 10 videos, got %d", config.Videos)
	}

	// 测试获取所有配置
	configs := GetAllBoostPackConfigs()
	if len(configs) != 3 {
		t.Errorf("Expected 3 boost pack configs, got %d", len(configs))
	}
}
