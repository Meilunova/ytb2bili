package membership

import (
	"testing"
	"time"
)

func TestTierConfig(t *testing.T) {
	// 测试所有等级配置存在
	for _, tier := range AllTiers {
		config := GetTierConfig(tier)
		if config.Tier != tier {
			t.Errorf("GetTierConfig(%s) returned wrong tier: %s", tier, config.Tier)
		}
	}
}

func TestUserMembershipIsExpired(t *testing.T) {
	tests := []struct {
		name      string
		tier      Tier
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "free tier never expires",
			tier:      TierFree,
			expiresAt: time.Now().Add(-24 * time.Hour), // 过期时间已过
			want:      false,
		},
		{
			name:      "basic tier not expired",
			tier:      TierBasic,
			expiresAt: time.Now().Add(24 * time.Hour),
			want:      false,
		},
		{
			name:      "basic tier expired",
			tier:      TierBasic,
			expiresAt: time.Now().Add(-24 * time.Hour),
			want:      true,
		},
		{
			name:      "pro tier not expired",
			tier:      TierPro,
			expiresAt: time.Now().Add(30 * 24 * time.Hour),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &UserMembership{
				UserID:    "test-user",
				Tier:      tt.tier,
				ExpiresAt: tt.expiresAt,
			}
			if got := m.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserMembershipGetEffectiveTier(t *testing.T) {
	tests := []struct {
		name      string
		tier      Tier
		expiresAt time.Time
		want      Tier
	}{
		{
			name:      "free tier stays free",
			tier:      TierFree,
			expiresAt: time.Time{},
			want:      TierFree,
		},
		{
			name:      "active basic tier",
			tier:      TierBasic,
			expiresAt: time.Now().Add(24 * time.Hour),
			want:      TierBasic,
		},
		{
			name:      "expired basic tier becomes free",
			tier:      TierBasic,
			expiresAt: time.Now().Add(-24 * time.Hour),
			want:      TierFree,
		},
		{
			name:      "active pro tier",
			tier:      TierPro,
			expiresAt: time.Now().Add(30 * 24 * time.Hour),
			want:      TierPro,
		},
		{
			name:      "expired enterprise tier becomes free",
			tier:      TierEnterprise,
			expiresAt: time.Now().Add(-1 * time.Hour),
			want:      TierFree,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &UserMembership{
				UserID:    "test-user",
				Tier:      tt.tier,
				ExpiresAt: tt.expiresAt,
			}
			if got := m.GetEffectiveTier(); got != tt.want {
				t.Errorf("GetEffectiveTier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserMembershipGetConfig(t *testing.T) {
	m := &UserMembership{
		UserID:    "test-user",
		Tier:      TierPro,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}

	config := m.GetConfig()
	if config.Tier != TierPro {
		t.Errorf("GetConfig() returned wrong tier: %s, want %s", config.Tier, TierPro)
	}
	if !config.Features.GeminiVideoAnalysis {
		t.Error("Pro tier should have GeminiVideoAnalysis enabled")
	}
	if config.Limits.VideosPerDay != 100 {
		t.Errorf("Pro tier should have 100 videos per day, got %d", config.Limits.VideosPerDay)
	}
}

func TestCompareTiers(t *testing.T) {
	tests := []struct {
		a, b Tier
		want int
	}{
		{TierFree, TierBasic, -1},
		{TierBasic, TierFree, 1},
		{TierBasic, TierBasic, 0},
		{TierPro, TierBasic, 1},
		{TierEnterprise, TierPro, 1},
		{TierFree, TierEnterprise, -1},
	}

	for _, tt := range tests {
		t.Run(string(tt.a)+"_vs_"+string(tt.b), func(t *testing.T) {
			if got := CompareTiers(tt.a, tt.b); got != tt.want {
				t.Errorf("CompareTiers(%s, %s) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsHigherTier(t *testing.T) {
	if !IsHigherTier(TierPro, TierBasic) {
		t.Error("Pro should be higher than Basic")
	}
	if IsHigherTier(TierBasic, TierPro) {
		t.Error("Basic should not be higher than Pro")
	}
	if IsHigherTier(TierBasic, TierBasic) {
		t.Error("Same tier should not be higher")
	}
}

func TestDaysUntilExpiry(t *testing.T) {
	// 免费用户永不过期
	free := &UserMembership{Tier: TierFree}
	if days := free.DaysUntilExpiry(); days != -1 {
		t.Errorf("Free tier should return -1, got %d", days)
	}

	// 已过期用户
	expired := &UserMembership{
		Tier:      TierBasic,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	if days := expired.DaysUntilExpiry(); days != 0 {
		t.Errorf("Expired tier should return 0, got %d", days)
	}

	// 还有30天过期
	active := &UserMembership{
		Tier:      TierPro,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	days := active.DaysUntilExpiry()
	if days < 29 || days > 31 {
		t.Errorf("Expected ~30 days, got %d", days)
	}
}

func TestFeaturesByTier(t *testing.T) {
	// 免费用户没有任何高级功能
	freeConfig := GetTierConfig(TierFree)
	if freeConfig.Features.AITranslation {
		t.Error("Free tier should not have AITranslation")
	}
	if freeConfig.Features.GeminiVideoAnalysis {
		t.Error("Free tier should not have GeminiVideoAnalysis")
	}

	// 基础版有 AI 翻译
	basicConfig := GetTierConfig(TierBasic)
	if !basicConfig.Features.AITranslation {
		t.Error("Basic tier should have AITranslation")
	}
	if basicConfig.Features.GeminiVideoAnalysis {
		t.Error("Basic tier should not have GeminiVideoAnalysis")
	}

	// 专业版有 Gemini 分析
	proConfig := GetTierConfig(TierPro)
	if !proConfig.Features.AITranslation {
		t.Error("Pro tier should have AITranslation")
	}
	if !proConfig.Features.GeminiVideoAnalysis {
		t.Error("Pro tier should have GeminiVideoAnalysis")
	}
	if proConfig.Features.APIAccess {
		t.Error("Pro tier should not have APIAccess")
	}

	// 企业版有所有功能
	enterpriseConfig := GetTierConfig(TierEnterprise)
	if !enterpriseConfig.Features.APIAccess {
		t.Error("Enterprise tier should have APIAccess")
	}
	if !enterpriseConfig.Features.TeamCollaboration {
		t.Error("Enterprise tier should have TeamCollaboration")
	}
	if enterpriseConfig.Limits.VideosPerDay != -1 {
		t.Error("Enterprise tier should have unlimited videos")
	}
}

func TestGetAllTierConfigs(t *testing.T) {
	configs := GetAllTierConfigs()
	if len(configs) != len(AllTiers) {
		t.Errorf("Expected %d configs, got %d", len(AllTiers), len(configs))
	}

	// 验证顺序
	for i, config := range configs {
		if config.Tier != AllTiers[i] {
			t.Errorf("Config at index %d has tier %s, expected %s", i, config.Tier, AllTiers[i])
		}
	}
}
