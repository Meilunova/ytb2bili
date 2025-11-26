package services

import (
	"testing"

	"github.com/difyz9/ytb2bili/internal/core/types"
	"go.uber.org/zap"
)

// 创建测试用的 logger
func createTestLogger() *zap.SugaredLogger {
	logger, _ := zap.NewDevelopment()
	return logger.Sugar()
}

// TestAIServicePriority 测试 AI 服务优先级
func TestAIServicePriority(t *testing.T) {
	logger := createTestLogger()

	t.Run("用户选择自定义API为首选", func(t *testing.T) {
		// 配置：启用自定义API和DeepSeek，用户选择自定义API
		config := &types.AppConfig{
			PrimaryAIService: "openai_compatible",
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled:  true,
				Provider: "custom",
				ApiKey:   "test-api-key-custom",
				BaseURL:  "https://api.example.com/v1",
				Model:    "gpt-4",
			},
			DeepSeekTransConfig: &types.DeepSeekTransConfig{
				Enabled:  true,
				ApiKey:   "test-api-key-deepseek",
				Model:    "deepseek-chat",
				Endpoint: "https://api.deepseek.com",
			},
		}

		manager := NewAIServiceManager(config, logger)

		// 获取首选提供商
		provider, err := manager.GetPreferredProvider()
		if err != nil {
			t.Fatalf("获取首选提供商失败: %v", err)
		}

		if provider != AIProviderOpenAICompatible {
			t.Errorf("期望首选提供商为 openai_compatible，实际为 %s", provider)
		}

		t.Logf("✅ 用户选择自定义API为首选: %s", provider)
	})

	t.Run("用户切换为DeepSeek为首选", func(t *testing.T) {
		// 配置：启用自定义API和DeepSeek，用户选择DeepSeek
		config := &types.AppConfig{
			PrimaryAIService: "deepseek", // 用户选择 DeepSeek
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled:  true,
				Provider: "custom",
				ApiKey:   "test-api-key-custom",
				BaseURL:  "https://api.example.com/v1",
				Model:    "gpt-4",
			},
			DeepSeekTransConfig: &types.DeepSeekTransConfig{
				Enabled:  true,
				ApiKey:   "test-api-key-deepseek",
				Model:    "deepseek-chat",
				Endpoint: "https://api.deepseek.com",
			},
		}

		manager := NewAIServiceManager(config, logger)

		// 获取首选提供商
		provider, err := manager.GetPreferredProvider()
		if err != nil {
			t.Fatalf("获取首选提供商失败: %v", err)
		}

		if provider != AIProviderDeepSeek {
			t.Errorf("期望首选提供商为 deepseek，实际为 %s", provider)
		}

		t.Logf("✅ 用户切换为DeepSeek为首选: %s", provider)
	})

	t.Run("动态切换首选服务", func(t *testing.T) {
		// 初始配置：用户选择自定义API
		config := &types.AppConfig{
			PrimaryAIService: "openai_compatible",
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled:  true,
				Provider: "custom",
				ApiKey:   "test-api-key-custom",
				BaseURL:  "https://api.example.com/v1",
				Model:    "gpt-4",
			},
			DeepSeekTransConfig: &types.DeepSeekTransConfig{
				Enabled:  true,
				ApiKey:   "test-api-key-deepseek",
				Model:    "deepseek-chat",
				Endpoint: "https://api.deepseek.com",
			},
		}

		manager := NewAIServiceManager(config, logger)

		// 验证初始首选为自定义API
		provider1, _ := manager.GetPreferredProvider()
		if provider1 != AIProviderOpenAICompatible {
			t.Errorf("初始首选应为 openai_compatible，实际为 %s", provider1)
		}
		t.Logf("✅ 初始首选服务: %s", provider1)

		// 模拟用户切换首选服务为 DeepSeek
		config.PrimaryAIService = "deepseek"
		manager.RefreshConfig(config)

		// 验证切换后首选为 DeepSeek
		provider2, _ := manager.GetPreferredProvider()
		if provider2 != AIProviderDeepSeek {
			t.Errorf("切换后首选应为 deepseek，实际为 %s", provider2)
		}
		t.Logf("✅ 切换后首选服务: %s", provider2)

		// 再次切换回自定义API
		config.PrimaryAIService = "openai_compatible"
		manager.RefreshConfig(config)

		provider3, _ := manager.GetPreferredProvider()
		if provider3 != AIProviderOpenAICompatible {
			t.Errorf("再次切换后首选应为 openai_compatible，实际为 %s", provider3)
		}
		t.Logf("✅ 再次切换后首选服务: %s", provider3)
	})

	t.Run("首选服务未启用时自动回退", func(t *testing.T) {
		// 配置：用户选择自定义API，但自定义API未启用
		config := &types.AppConfig{
			PrimaryAIService: "openai_compatible",
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled: false, // 未启用
			},
			DeepSeekTransConfig: &types.DeepSeekTransConfig{
				Enabled:  true,
				ApiKey:   "test-api-key-deepseek",
				Model:    "deepseek-chat",
				Endpoint: "https://api.deepseek.com",
			},
		}

		manager := NewAIServiceManager(config, logger)

		// 应该自动回退到 DeepSeek
		provider, err := manager.GetPreferredProvider()
		if err != nil {
			t.Fatalf("获取首选提供商失败: %v", err)
		}

		if provider != AIProviderDeepSeek {
			t.Errorf("首选服务未启用时应回退到 deepseek，实际为 %s", provider)
		}

		t.Logf("✅ 首选服务未启用，自动回退到: %s", provider)
	})

	t.Run("用户选择Gemini为首选", func(t *testing.T) {
		// 配置：启用所有服务，用户选择Gemini
		config := &types.AppConfig{
			PrimaryAIService: "gemini",
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled:  true,
				Provider: "custom",
				ApiKey:   "test-api-key-custom",
				BaseURL:  "https://api.example.com/v1",
				Model:    "gpt-4",
			},
			DeepSeekTransConfig: &types.DeepSeekTransConfig{
				Enabled:  true,
				ApiKey:   "test-api-key-deepseek",
				Model:    "deepseek-chat",
				Endpoint: "https://api.deepseek.com",
			},
			GeminiConfig: &types.GeminiConfig{
				Enabled: true,
				ApiKey:  "test-api-key-gemini",
				Model:   "gemini-pro",
			},
		}

		manager := NewAIServiceManager(config, logger)

		provider, err := manager.GetPreferredProvider()
		if err != nil {
			t.Fatalf("获取首选提供商失败: %v", err)
		}

		if provider != AIProviderGemini {
			t.Errorf("期望首选提供商为 gemini，实际为 %s", provider)
		}

		t.Logf("✅ 用户选择Gemini为首选: %s", provider)
	})

	t.Run("未设置首选时按默认优先级", func(t *testing.T) {
		// 配置：未设置首选服务，启用所有服务
		config := &types.AppConfig{
			PrimaryAIService: "", // 未设置
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled:  true,
				Provider: "custom",
				ApiKey:   "test-api-key-custom",
				BaseURL:  "https://api.example.com/v1",
				Model:    "gpt-4",
			},
			DeepSeekTransConfig: &types.DeepSeekTransConfig{
				Enabled:  true,
				ApiKey:   "test-api-key-deepseek",
				Model:    "deepseek-chat",
				Endpoint: "https://api.deepseek.com",
			},
			GeminiConfig: &types.GeminiConfig{
				Enabled: true,
				ApiKey:  "test-api-key-gemini",
				Model:   "gemini-pro",
			},
		}

		manager := NewAIServiceManager(config, logger)

		// 未设置首选时，应按默认优先级选择 openai_compatible
		provider, err := manager.GetPreferredProvider()
		if err != nil {
			t.Fatalf("获取首选提供商失败: %v", err)
		}

		if provider != AIProviderOpenAICompatible {
			t.Errorf("未设置首选时应按默认优先级选择 openai_compatible，实际为 %s", provider)
		}

		t.Logf("✅ 未设置首选时按默认优先级: %s", provider)
	})

	t.Run("获取服务状态", func(t *testing.T) {
		config := &types.AppConfig{
			PrimaryAIService: "openai_compatible",
			OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
				Enabled:  true,
				Provider: "gemini",
				ApiKey:   "test-api-key",
				BaseURL:  "https://api.example.com/v1",
				Model:    "gemini-2.5-pro",
			},
		}

		manager := NewAIServiceManager(config, logger)

		status := manager.GetStatus(AIProviderOpenAICompatible)
		if status == nil {
			t.Fatal("获取状态失败")
		}

		if !status.Enabled {
			t.Error("服务应该是启用状态")
		}

		if status.Model != "gemini-2.5-pro" {
			t.Errorf("模型应为 gemini-2.5-pro，实际为 %s", status.Model)
		}

		t.Logf("✅ 服务状态: Name=%s, Model=%s, Enabled=%v", status.Name, status.Model, status.Enabled)
	})
}

// TestAIServiceManagerMethods 测试 AI 服务管理器的其他方法
func TestAIServiceManagerMethods(t *testing.T) {
	logger := createTestLogger()

	config := &types.AppConfig{
		PrimaryAIService: "openai_compatible",
		OpenAICompatibleConfig: &types.OpenAICompatibleConfig{
			Enabled:  true,
			Provider: "custom",
			ApiKey:   "test-key",
			BaseURL:  "https://api.example.com/v1",
			Model:    "gpt-4",
		},
		DeepSeekTransConfig: &types.DeepSeekTransConfig{
			Enabled:  true,
			ApiKey:   "test-key",
			Model:    "deepseek-chat",
			Endpoint: "https://api.deepseek.com",
		},
	}

	manager := NewAIServiceManager(config, logger)

	t.Run("IsOpenAICompatibleEnabled", func(t *testing.T) {
		if !manager.IsOpenAICompatibleEnabled() {
			t.Error("OpenAI兼容API应该是启用状态")
		}
		t.Log("✅ IsOpenAICompatibleEnabled 正确")
	})

	t.Run("IsDeepSeekEnabled", func(t *testing.T) {
		if !manager.IsDeepSeekEnabled() {
			t.Error("DeepSeek应该是启用状态")
		}
		t.Log("✅ IsDeepSeekEnabled 正确")
	})

	t.Run("GetOpenAICompatibleConfig", func(t *testing.T) {
		cfg := manager.GetOpenAICompatibleConfig()
		if cfg == nil {
			t.Fatal("获取配置失败")
		}
		if cfg.Model != "gpt-4" {
			t.Errorf("模型应为 gpt-4，实际为 %s", cfg.Model)
		}
		t.Logf("✅ GetOpenAICompatibleConfig: Model=%s", cfg.Model)
	})

	t.Run("GetDeepSeekConfig", func(t *testing.T) {
		cfg := manager.GetDeepSeekConfig()
		if cfg == nil {
			t.Fatal("获取配置失败")
		}
		if cfg.Model != "deepseek-chat" {
			t.Errorf("模型应为 deepseek-chat，实际为 %s", cfg.Model)
		}
		t.Logf("✅ GetDeepSeekConfig: Model=%s", cfg.Model)
	})

	t.Run("GetAllStatus", func(t *testing.T) {
		statuses := manager.GetAllStatus()
		if len(statuses) == 0 {
			t.Error("应该返回服务状态列表")
		}

		for _, status := range statuses {
			t.Logf("  - %s: Enabled=%v, Model=%s", status.Name, status.Enabled, status.Model)
		}
		t.Log("✅ GetAllStatus 正确")
	})
}
