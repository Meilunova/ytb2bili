package membership

import (
	"context"
	"testing"
	"time"
)

func TestQuotaService_GetQuotaInfo(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	// 免费用户
	store.SetUserTier("free-user", TierFree)
	info, err := quotaService.GetQuotaInfo(ctx, "free-user")
	if err != nil {
		t.Fatalf("GetQuotaInfo failed: %v", err)
	}
	if info.DailyLimit != 5 {
		t.Errorf("Free user daily limit should be 5, got %d", info.DailyLimit)
	}
	if info.IsUnlimited {
		t.Error("Free user should not be unlimited")
	}
	if info.DailyRemaining != 5 {
		t.Errorf("Free user daily remaining should be 5, got %d", info.DailyRemaining)
	}

	// 企业版用户（无限）
	store.SetUserTier("enterprise-user", TierEnterprise)
	info, err = quotaService.GetQuotaInfo(ctx, "enterprise-user")
	if err != nil {
		t.Fatalf("GetQuotaInfo failed: %v", err)
	}
	if !info.IsUnlimited {
		t.Error("Enterprise user should be unlimited")
	}
	if info.DailyLimit != -1 {
		t.Errorf("Enterprise user daily limit should be -1, got %d", info.DailyLimit)
	}
}

func TestQuotaService_GetQuotaInfo_WithUsage(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("user1", TierFree)
	today := time.Now().Format("2006-01-02")
	store.SetDailyUsage("user1", today, 3)

	info, _ := quotaService.GetQuotaInfo(ctx, "user1")
	if info.DailyUsed != 3 {
		t.Errorf("Daily used should be 3, got %d", info.DailyUsed)
	}
	if info.DailyRemaining != 2 {
		t.Errorf("Daily remaining should be 2, got %d", info.DailyRemaining)
	}
}

func TestQuotaService_GetQuotaInfo_WithBoostPack(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("user1", TierFree)
	store.SetBoostPack("user1", 10, 7)

	info, _ := quotaService.GetQuotaInfo(ctx, "user1")
	if info.BoostPackRemaining != 10 {
		t.Errorf("Boost pack remaining should be 10, got %d", info.BoostPackRemaining)
	}
	if info.TotalRemaining != 15 { // 5 daily + 10 boost
		t.Errorf("Total remaining should be 15, got %d", info.TotalRemaining)
	}
}

func TestQuotaService_ConsumeQuota(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("user1", TierFree)
	today := time.Now().Format("2006-01-02")

	// 消耗配额
	for i := 0; i < 5; i++ {
		err := quotaService.ConsumeQuota(ctx, "user1")
		if err != nil {
			t.Errorf("ConsumeQuota %d failed: %v", i+1, err)
		}
	}

	// 验证使用量
	usage, _ := store.GetDailyUsage(ctx, "user1", today)
	if usage != 5 {
		t.Errorf("Usage should be 5, got %d", usage)
	}

	// 第6次应该失败
	err := quotaService.ConsumeQuota(ctx, "user1")
	if err == nil {
		t.Error("6th consume should fail")
	}
}

func TestQuotaService_ConsumeQuota_WithBoostPack(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("user1", TierFree)
	today := time.Now().Format("2006-01-02")
	store.SetDailyUsage("user1", today, 5) // 每日配额用完
	store.SetBoostPack("user1", 3, 7)      // 加油包3个

	// 应该消耗加油包
	err := quotaService.ConsumeQuota(ctx, "user1")
	if err != nil {
		t.Errorf("ConsumeQuota with boost pack failed: %v", err)
	}

	// 验证加油包减少
	pack, _ := store.GetUserBoostPack(ctx, "user1")
	if pack.VideosRemaining != 2 {
		t.Errorf("Boost pack should have 2 remaining, got %d", pack.VideosRemaining)
	}
}

func TestQuotaService_ConsumeQuota_Unlimited(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("enterprise-user", TierEnterprise)

	// 无限用户消耗不应该失败
	for i := 0; i < 100; i++ {
		err := quotaService.ConsumeQuota(ctx, "enterprise-user")
		if err != nil {
			t.Errorf("Enterprise user consume %d failed: %v", i+1, err)
		}
	}
}

func TestQuotaService_HasQuota(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("user1", TierFree)

	// 初始有配额
	if !quotaService.HasQuota(ctx, "user1") {
		t.Error("User should have quota initially")
	}

	// 用完配额
	today := time.Now().Format("2006-01-02")
	store.SetDailyUsage("user1", today, 5)

	if quotaService.HasQuota(ctx, "user1") {
		t.Error("User should not have quota after using all")
	}

	// 添加加油包
	store.SetBoostPack("user1", 1, 7)
	if !quotaService.HasQuota(ctx, "user1") {
		t.Error("User should have quota with boost pack")
	}
}

func TestQuotaService_GetBatchLimit(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	tests := []struct {
		tier  Tier
		limit int
	}{
		{TierFree, 1},
		{TierBasic, 5},
		{TierPro, 20},
		{TierEnterprise, 100},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			userID := "user-" + string(tt.tier)
			store.SetUserTier(userID, tt.tier)

			limit := quotaService.GetBatchLimit(ctx, userID)
			if limit != tt.limit {
				t.Errorf("GetBatchLimit(%s) = %d, want %d", tt.tier, limit, tt.limit)
			}
		})
	}
}

func TestQuotaService_CanBatch(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	quotaService := NewQuotaService(store, checker)
	ctx := context.Background()

	store.SetUserTier("basic-user", TierBasic)

	if !quotaService.CanBatch(ctx, "basic-user", 5) {
		t.Error("Basic user should be able to batch 5")
	}
	if quotaService.CanBatch(ctx, "basic-user", 6) {
		t.Error("Basic user should not be able to batch 6")
	}
}
