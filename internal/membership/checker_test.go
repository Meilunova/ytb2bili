package membership

import (
	"context"
	"testing"
	"time"
)

func TestFeatureChecker_CanUseFeature(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	tests := []struct {
		name    string
		tier    Tier
		feature string
		want    bool
	}{
		// 免费用户
		{"free_ai_translation", TierFree, "ai_translation", false},
		{"free_gemini", TierFree, "gemini_video_analysis", false},
		{"free_api_access", TierFree, "api_access", false},

		// 基础版用户
		{"basic_ai_translation", TierBasic, "ai_translation", true},
		{"basic_ai_title", TierBasic, "ai_title_generation", true},
		{"basic_gemini", TierBasic, "gemini_video_analysis", false},
		{"basic_api_access", TierBasic, "api_access", false},

		// 专业版用户
		{"pro_ai_translation", TierPro, "ai_translation", true},
		{"pro_gemini", TierPro, "gemini_video_analysis", true},
		{"pro_auto_upload", TierPro, "auto_upload", true},
		{"pro_api_access", TierPro, "api_access", false},

		// 企业版用户
		{"enterprise_api_access", TierEnterprise, "api_access", true},
		{"enterprise_team", TierEnterprise, "team_collaboration", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := "test-user-" + tt.name
			store.SetUserTier(userID, tt.tier)

			result := checker.CanUseFeature(ctx, userID, tt.feature)
			if result.Allowed != tt.want {
				t.Errorf("CanUseFeature(%s, %s) = %v, want %v", tt.tier, tt.feature, result.Allowed, tt.want)
			}

			// 检查拒绝时的升级建议
			if !result.Allowed && result.Upgrade == "" {
				t.Errorf("Expected upgrade suggestion when feature not allowed")
			}
		})
	}
}

func TestFeatureChecker_CanUseFeature_UnknownFeature(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	store.SetUserTier("user1", TierFree)
	result := checker.CanUseFeature(ctx, "user1", "unknown_feature")

	// 未知功能默认允许
	if !result.Allowed {
		t.Error("Unknown feature should be allowed by default")
	}
}

func TestFeatureChecker_CanProcessVideo(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	// 免费用户，每日5个配额
	store.SetUserTier("free-user", TierFree)

	// 前5个应该允许
	for i := 0; i < 5; i++ {
		result := checker.CanProcessVideo(ctx, "free-user")
		if !result.Allowed {
			t.Errorf("Video %d should be allowed", i+1)
		}
		store.IncrDailyUsage(ctx, "free-user", time.Now().Format("2006-01-02"))
	}

	// 第6个应该被拒绝
	result := checker.CanProcessVideo(ctx, "free-user")
	if result.Allowed {
		t.Error("Video 6 should be denied for free user")
	}
	if result.Code != "QUOTA_EXCEEDED" {
		t.Errorf("Expected QUOTA_EXCEEDED code, got %s", result.Code)
	}
}

func TestFeatureChecker_CanProcessVideo_WithBoostPack(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	// 免费用户，配额用完
	store.SetUserTier("user1", TierFree)
	store.SetDailyUsage("user1", time.Now().Format("2006-01-02"), 5)

	// 没有加油包，应该被拒绝
	result := checker.CanProcessVideo(ctx, "user1")
	if result.Allowed {
		t.Error("Should be denied without boost pack")
	}

	// 添加加油包
	store.SetBoostPack("user1", 10, 7)

	// 有加油包，应该允许
	result = checker.CanProcessVideo(ctx, "user1")
	if !result.Allowed {
		t.Error("Should be allowed with boost pack")
	}
}

func TestFeatureChecker_CanProcessVideo_Unlimited(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	// 企业版用户，无限配额
	store.SetUserTier("enterprise-user", TierEnterprise)

	// 即使使用了很多，也应该允许
	store.SetDailyUsage("enterprise-user", time.Now().Format("2006-01-02"), 1000)

	result := checker.CanProcessVideo(ctx, "enterprise-user")
	if !result.Allowed {
		t.Error("Enterprise user should have unlimited quota")
	}
}

func TestFeatureChecker_CanBatchSubmit(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	tests := []struct {
		name  string
		tier  Tier
		count int
		want  bool
	}{
		{"free_1", TierFree, 1, true},
		{"free_2", TierFree, 2, false}, // 免费版只能1个
		{"basic_5", TierBasic, 5, true},
		{"basic_6", TierBasic, 6, false}, // 基础版最多5个
		{"pro_20", TierPro, 20, true},
		{"pro_21", TierPro, 21, false}, // 专业版最多20个
		{"enterprise_100", TierEnterprise, 100, true},
		{"enterprise_101", TierEnterprise, 101, false}, // 企业版最多100个
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := "user-" + tt.name
			store.SetUserTier(userID, tt.tier)

			result := checker.CanBatchSubmit(ctx, userID, tt.count)
			if result.Allowed != tt.want {
				t.Errorf("CanBatchSubmit(%s, %d) = %v, want %v", tt.tier, tt.count, result.Allowed, tt.want)
			}
		})
	}
}

func TestFeatureChecker_GetUserPriority(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	tests := []struct {
		tier     Tier
		priority int
	}{
		{TierFree, 0},
		{TierBasic, 1},
		{TierPro, 2},
		{TierEnterprise, 3},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			userID := "user-" + string(tt.tier)
			store.SetUserTier(userID, tt.tier)

			priority := checker.GetUserPriority(ctx, userID)
			if priority != tt.priority {
				t.Errorf("GetUserPriority(%s) = %d, want %d", tt.tier, priority, tt.priority)
			}
		})
	}
}

func TestFeatureChecker_GetUserTier(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	// 新用户应该是免费
	tier := checker.GetUserTier(ctx, "new-user")
	if tier != TierFree {
		t.Errorf("New user should be free tier, got %s", tier)
	}

	// 设置为专业版
	store.SetUserTier("pro-user", TierPro)
	tier = checker.GetUserTier(ctx, "pro-user")
	if tier != TierPro {
		t.Errorf("Expected Pro tier, got %s", tier)
	}

	// 过期的专业版应该返回免费
	store.SetUserTierWithExpiry("expired-user", TierPro, time.Now().Add(-24*time.Hour))
	tier = checker.GetUserTier(ctx, "expired-user")
	if tier != TierFree {
		t.Errorf("Expired user should be free tier, got %s", tier)
	}
}

func TestFeatureChecker_SuggestUpgrade(t *testing.T) {
	store := NewMockStore()
	checker := NewFeatureChecker(store)
	ctx := context.Background()

	// 免费用户请求付费功能，应该建议升级到基础版
	store.SetUserTier("free-user", TierFree)
	result := checker.CanUseFeature(ctx, "free-user", "ai_translation")
	if result.Upgrade != string(TierBasic) {
		t.Errorf("Free user should be suggested to upgrade to Basic, got %s", result.Upgrade)
	}

	// 基础版用户请求专业版功能，应该建议升级到专业版
	store.SetUserTier("basic-user", TierBasic)
	result = checker.CanUseFeature(ctx, "basic-user", "gemini_video_analysis")
	if result.Upgrade != string(TierPro) {
		t.Errorf("Basic user should be suggested to upgrade to Pro, got %s", result.Upgrade)
	}

	// 专业版用户请求企业版功能，应该建议升级到企业版
	store.SetUserTier("pro-user", TierPro)
	result = checker.CanUseFeature(ctx, "pro-user", "api_access")
	if result.Upgrade != string(TierEnterprise) {
		t.Errorf("Pro user should be suggested to upgrade to Enterprise, got %s", result.Upgrade)
	}
}
