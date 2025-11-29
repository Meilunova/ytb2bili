package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/difyz9/ytb2bili/internal/core"
	"github.com/difyz9/ytb2bili/internal/core/types"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	BaseHandler
}

func NewConfigHandler(app *core.AppServer) *ConfigHandler {
	return &ConfigHandler{
		BaseHandler: BaseHandler{App: app},
	}
}

// RegisterRoutes 注册配置相关路由
func (h *ConfigHandler) RegisterRoutes(server *core.AppServer) {
	api := server.Engine.Group("/api/v1")

	config := api.Group("/config")
	{
		config.GET("/deepseek", h.getDeepSeekConfig)
		config.PUT("/deepseek", h.updateDeepSeekConfig)
		config.GET("/proxy", h.getProxyConfig)
		config.PUT("/proxy", h.updateProxyConfig)

		// OpenAI兼容API配置
		config.GET("/openai-compatible", h.getOpenAICompatibleConfig)
		config.PUT("/openai-compatible", h.updateOpenAICompatibleConfig)
		config.POST("/openai-compatible/test", h.testOpenAICompatibleAPI)
		config.GET("/openai-compatible/providers", h.getOpenAICompatibleProviders)

		// AI服务状态
		config.GET("/ai-services/status", h.getAIServicesStatus)
		config.PUT("/ai-services/primary", h.setPrimaryAIService)

		// Gemini原生配置（用于元数据生成）
		config.GET("/gemini", h.getGeminiConfig)
		config.PUT("/gemini", h.updateGeminiConfig)
		config.POST("/gemini/validate", h.validateGeminiApiKeys)
		config.GET("/gemini/models", h.getGeminiModels)
	}
}

// DeepSeekConfigRequest DeepSeek配置请求
type DeepSeekConfigRequest struct {
	Enabled   *bool   `json:"enabled,omitempty"`    // 是否启用（可选）
	ApiKey    *string `json:"api_key,omitempty"`    // API Key（可选）
	Model     *string `json:"model,omitempty"`      // 模型（可选）
	Endpoint  *string `json:"endpoint,omitempty"`   // 端点（可选）
	Timeout   *int    `json:"timeout,omitempty"`    // 超时时间（可选）
	MaxTokens *int    `json:"max_tokens,omitempty"` // 最大Token数（可选）
}

// DeepSeekConfigResponse DeepSeek配置响应
type DeepSeekConfigResponse struct {
	Enabled   bool   `json:"enabled"`
	ApiKey    string `json:"api_key"` // 为了安全只返回部分字符
	Model     string `json:"model"`
	Endpoint  string `json:"endpoint"`
	Timeout   int    `json:"timeout"`
	MaxTokens int    `json:"max_tokens"`
}

// ProxyConfigRequest 代理配置请求
type ProxyConfigRequest struct {
	UseProxy  *bool   `json:"useProxy,omitempty"`  // 是否使用代理（可选）
	ProxyHost *string `json:"proxyHost,omitempty"` // 代理地址（可选）
}

// ProxyConfigResponse 代理配置响应
type ProxyConfigResponse struct {
	UseProxy  bool   `json:"useProxy"`  // 是否使用代理
	ProxyHost string `json:"proxyHost"` // 代理地址
}

// getDeepSeekConfig 获取DeepSeek配置
func (h *ConfigHandler) getDeepSeekConfig(c *gin.Context) {
	config := h.App.Config.DeepSeekTransConfig
	if config == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data": DeepSeekConfigResponse{
				Enabled:   false,
				ApiKey:    "",
				Model:     "deepseek-chat",
				Endpoint:  "https://api.deepseek.com",
				Timeout:   60,
				MaxTokens: 4000,
			},
		})
		return
	}

	// 隐藏完整的API Key，只显示前几位和后几位
	apiKeyMasked := ""
	if config.ApiKey != "" {
		if len(config.ApiKey) > 10 {
			apiKeyMasked = config.ApiKey[:6] + "..." + config.ApiKey[len(config.ApiKey)-4:]
		} else {
			apiKeyMasked = "***"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": DeepSeekConfigResponse{
			Enabled:   config.Enabled,
			ApiKey:    apiKeyMasked,
			Model:     config.Model,
			Endpoint:  config.Endpoint,
			Timeout:   config.Timeout,
			MaxTokens: config.MaxTokens,
		},
	})
}

// updateDeepSeekConfig 更新DeepSeek配置
func (h *ConfigHandler) updateDeepSeekConfig(c *gin.Context) {
	var req DeepSeekConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 确保配置对象存在
	if h.App.Config.DeepSeekTransConfig == nil {
		h.App.Config.DeepSeekTransConfig = &types.DeepSeekTransConfig{
			Enabled:   false,
			ApiKey:    "",
			Model:     "deepseek-chat",
			Endpoint:  "https://api.deepseek.com",
			Timeout:   60,
			MaxTokens: 4000,
		}
	}

	config := h.App.Config.DeepSeekTransConfig

	// 更新配置字段（只更新提供的字段）
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
		h.App.Logger.Infof("Updated DeepSeek enabled: %v", config.Enabled)
	}

	if req.ApiKey != nil {
		config.ApiKey = *req.ApiKey
		h.App.Logger.Infof("Updated DeepSeek API Key: %s", maskApiKey(*req.ApiKey))
	}

	if req.Model != nil {
		config.Model = *req.Model
		h.App.Logger.Infof("Updated DeepSeek model: %s", config.Model)
	}

	if req.Endpoint != nil {
		config.Endpoint = *req.Endpoint
		h.App.Logger.Infof("Updated DeepSeek endpoint: %s", config.Endpoint)
	}

	if req.Timeout != nil {
		config.Timeout = *req.Timeout
		h.App.Logger.Infof("Updated DeepSeek timeout: %d", config.Timeout)
	}

	if req.MaxTokens != nil {
		config.MaxTokens = *req.MaxTokens
		h.App.Logger.Infof("Updated DeepSeek max_tokens: %d", config.MaxTokens)
	}

	// 保存配置到文件
	if err := types.SaveConfig(h.App.Config); err != nil {
		h.App.Logger.Errorf("Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to save configuration: " + err.Error(),
		})
		return
	}

	// 实时更新应用服务器的配置（不需要重启）
	h.App.Config.DeepSeekTransConfig = config
	h.App.Logger.Info("✅ DeepSeek configuration updated and applied successfully (no restart required)")

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Configuration updated and applied successfully (no restart required)",
		"data": DeepSeekConfigResponse{
			Enabled:   config.Enabled,
			ApiKey:    maskApiKey(config.ApiKey),
			Model:     config.Model,
			Endpoint:  config.Endpoint,
			Timeout:   config.Timeout,
			MaxTokens: config.MaxTokens,
		},
	})
}

// getProxyConfig 获取代理配置
func (h *ConfigHandler) getProxyConfig(c *gin.Context) {
	// 检查配置中是否有代理配置
	proxyConfig := h.App.Config.ProxyConfig
	if proxyConfig == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data": ProxyConfigResponse{
				UseProxy:  false,
				ProxyHost: "",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": ProxyConfigResponse{
			UseProxy:  proxyConfig.UseProxy,
			ProxyHost: proxyConfig.ProxyHost,
		},
	})
}

// updateProxyConfig 更新代理配置
func (h *ConfigHandler) updateProxyConfig(c *gin.Context) {
	var req ProxyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 确保配置对象存在
	if h.App.Config.ProxyConfig == nil {
		h.App.Config.ProxyConfig = &types.ProxyConfig{
			UseProxy:  false,
			ProxyHost: "",
		}
	}

	config := h.App.Config.ProxyConfig

	// 更新配置字段（只更新提供的字段）
	if req.UseProxy != nil {
		config.UseProxy = *req.UseProxy
		h.App.Logger.Infof("Updated proxy enabled: %v", config.UseProxy)
	}

	if req.ProxyHost != nil {
		config.ProxyHost = *req.ProxyHost
		h.App.Logger.Infof("Updated proxy host: %s", config.ProxyHost)
	}

	// 保存配置到文件
	if err := types.SaveConfig(h.App.Config); err != nil {
		h.App.Logger.Errorf("Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to save configuration: " + err.Error(),
		})
		return
	}

	// 实时更新应用服务器的配置（不需要重启）
	h.App.Config.ProxyConfig = config
	h.App.Logger.Info("✅ Proxy configuration updated and applied successfully (no restart required)")

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Configuration updated and applied successfully (no restart required)",
		"data": ProxyConfigResponse{
			UseProxy:  config.UseProxy,
			ProxyHost: config.ProxyHost,
		},
	})
}

// maskApiKey 隐藏API Key的敏感信息
func maskApiKey(apiKey string) string {
	if apiKey == "" {
		return ""
	}
	if len(apiKey) > 10 {
		return apiKey[:6] + "..." + apiKey[len(apiKey)-4:]
	}
	return "***"
}

// ========== OpenAI兼容API配置 ==========

// OpenAICompatibleConfigRequest OpenAI兼容API配置请求
type OpenAICompatibleConfigRequest struct {
	Enabled     *bool    `json:"enabled,omitempty"`
	Provider    *string  `json:"provider,omitempty"`
	ApiKey      *string  `json:"api_key,omitempty"`
	BaseURL     *string  `json:"base_url,omitempty"`
	Model       *string  `json:"model,omitempty"`
	Timeout     *int     `json:"timeout,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
}

// OpenAICompatibleConfigResponse OpenAI兼容API配置响应
type OpenAICompatibleConfigResponse struct {
	Enabled     bool    `json:"enabled"`
	Provider    string  `json:"provider"`
	ApiKey      string  `json:"api_key"` // 脱敏显示
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	Timeout     int     `json:"timeout"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// OpenAIProviderInfo 提供商信息
type OpenAIProviderInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	BaseURL      string `json:"base_url"`
	DefaultModel string `json:"default_model"`
}

// TestAPIRequest 测试API请求
type TestAPIRequest struct {
	Provider    string  `json:"provider"`
	ApiKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	Timeout     int     `json:"timeout"`
	Temperature float64 `json:"temperature"`
}

// TestAPIResponse 测试API响应
type TestAPIResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Response string `json:"response,omitempty"`
	Latency  int64  `json:"latency_ms,omitempty"`
}

// getOpenAICompatibleConfig 获取OpenAI兼容API配置
func (h *ConfigHandler) getOpenAICompatibleConfig(c *gin.Context) {
	config := h.App.Config.OpenAICompatibleConfig
	if config == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data": OpenAICompatibleConfigResponse{
				Enabled:     false,
				Provider:    "openai",
				ApiKey:      "",
				BaseURL:     "https://api.openai.com/v1",
				Model:       "gpt-3.5-turbo",
				Timeout:     60,
				MaxTokens:   4000,
				Temperature: 0.7,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": OpenAICompatibleConfigResponse{
			Enabled:     config.Enabled,
			Provider:    config.Provider,
			ApiKey:      maskApiKey(config.ApiKey),
			BaseURL:     config.BaseURL,
			Model:       config.Model,
			Timeout:     config.Timeout,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		},
	})
}

// updateOpenAICompatibleConfig 更新OpenAI兼容API配置
func (h *ConfigHandler) updateOpenAICompatibleConfig(c *gin.Context) {
	var req OpenAICompatibleConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 确保配置对象存在
	if h.App.Config.OpenAICompatibleConfig == nil {
		h.App.Config.OpenAICompatibleConfig = &types.OpenAICompatibleConfig{
			Enabled:     false,
			Provider:    "openai",
			ApiKey:      "",
			BaseURL:     "https://api.openai.com/v1",
			Model:       "gpt-3.5-turbo",
			Timeout:     60,
			MaxTokens:   4000,
			Temperature: 0.7,
		}
	}

	config := h.App.Config.OpenAICompatibleConfig

	// 更新配置字段
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
		h.App.Logger.Infof("Updated OpenAI Compatible enabled: %v", config.Enabled)
	}
	if req.Provider != nil {
		config.Provider = *req.Provider
		h.App.Logger.Infof("Updated OpenAI Compatible provider: %s", config.Provider)
	}
	if req.ApiKey != nil {
		config.ApiKey = *req.ApiKey
		h.App.Logger.Infof("Updated OpenAI Compatible API Key: %s", maskApiKey(*req.ApiKey))
	}
	if req.BaseURL != nil {
		config.BaseURL = *req.BaseURL
		h.App.Logger.Infof("Updated OpenAI Compatible base URL: %s", config.BaseURL)
	}
	if req.Model != nil {
		config.Model = *req.Model
		h.App.Logger.Infof("Updated OpenAI Compatible model: %s", config.Model)
	}
	if req.Timeout != nil {
		config.Timeout = *req.Timeout
		h.App.Logger.Infof("Updated OpenAI Compatible timeout: %d", config.Timeout)
	}
	if req.MaxTokens != nil {
		config.MaxTokens = *req.MaxTokens
		h.App.Logger.Infof("Updated OpenAI Compatible max_tokens: %d", config.MaxTokens)
	}
	if req.Temperature != nil {
		config.Temperature = *req.Temperature
		h.App.Logger.Infof("Updated OpenAI Compatible temperature: %f", config.Temperature)
	}

	// 保存配置到文件
	if err := types.SaveConfig(h.App.Config); err != nil {
		h.App.Logger.Errorf("Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to save configuration: " + err.Error(),
		})
		return
	}

	h.App.Config.OpenAICompatibleConfig = config
	h.App.Logger.Info("✅ OpenAI Compatible configuration updated successfully")

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Configuration updated successfully",
		"data": OpenAICompatibleConfigResponse{
			Enabled:     config.Enabled,
			Provider:    config.Provider,
			ApiKey:      maskApiKey(config.ApiKey),
			BaseURL:     config.BaseURL,
			Model:       config.Model,
			Timeout:     config.Timeout,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		},
	})
}

// testOpenAICompatibleAPI 测试OpenAI兼容API连接
func (h *ConfigHandler) testOpenAICompatibleAPI(c *gin.Context) {
	var req TestAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 验证必要参数
	if req.ApiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "API Key is required",
		})
		return
	}
	if req.BaseURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Base URL is required",
		})
		return
	}

	// 设置默认值
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}

	h.App.Logger.Infof("Testing OpenAI Compatible API: provider=%s, base_url=%s, model=%s",
		req.Provider, req.BaseURL, req.Model)

	// 调用测试
	testResult := h.doTestOpenAIAPI(req)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Test completed",
		"data":    testResult,
	})
}

// doTestOpenAIAPI 执行API测试
func (h *ConfigHandler) doTestOpenAIAPI(req TestAPIRequest) TestAPIResponse {
	startTime := time.Now()

	// 构建请求
	requestBody := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant."},
			{"role": "user", "content": "Say 'OK' if you can hear me."},
		},
		"max_tokens":  50,
		"temperature": req.Temperature,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return TestAPIResponse{
			Success: false,
			Message: "Failed to create request: " + err.Error(),
		}
	}

	// 构建URL
	apiURL := req.BaseURL
	if !strings.HasSuffix(apiURL, "/v1") && !strings.Contains(apiURL, "/chat/completions") {
		apiURL = strings.TrimSuffix(apiURL, "/") + "/v1"
	}
	if !strings.Contains(apiURL, "/chat/completions") {
		apiURL = strings.TrimSuffix(apiURL, "/") + "/chat/completions"
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return TestAPIResponse{
			Success: false,
			Message: "Failed to create HTTP request: " + err.Error(),
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.ApiKey)

	client := &http.Client{
		Timeout: time.Duration(req.Timeout) * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return TestAPIResponse{
			Success: false,
			Message: "API request failed: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TestAPIResponse{
			Success: false,
			Message: "Failed to read response: " + err.Error(),
		}
	}

	latency := time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		return TestAPIResponse{
			Success: false,
			Message: fmt.Sprintf("API returned error (status %d): %s", resp.StatusCode, string(body)),
			Latency: latency,
		}
	}

	// 解析响应
	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return TestAPIResponse{
			Success: false,
			Message: "Failed to parse response: " + err.Error(),
			Latency: latency,
		}
	}

	if apiResp.Error != nil {
		return TestAPIResponse{
			Success: false,
			Message: "API error: " + apiResp.Error.Message,
			Latency: latency,
		}
	}

	responseText := ""
	if len(apiResp.Choices) > 0 {
		responseText = apiResp.Choices[0].Message.Content
	}

	return TestAPIResponse{
		Success:  true,
		Message:  "API connection successful",
		Response: responseText,
		Latency:  latency,
	}
}

// getOpenAICompatibleProviders 获取支持的提供商列表
func (h *ConfigHandler) getOpenAICompatibleProviders(c *gin.Context) {
	providers := []OpenAIProviderInfo{
		{
			ID:           "openai",
			Name:         "OpenAI",
			Description:  "OpenAI官方API，支持GPT-3.5/GPT-4等模型",
			BaseURL:      "https://api.openai.com/v1",
			DefaultModel: "gpt-3.5-turbo",
		},
		{
			ID:           "deepseek",
			Name:         "DeepSeek",
			Description:  "DeepSeek深度求索，国产大模型",
			BaseURL:      "https://api.deepseek.com/v1",
			DefaultModel: "deepseek-chat",
		},
		{
			ID:           "qwen",
			Name:         "通义千问",
			Description:  "阿里云通义千问，支持OpenAI兼容模式",
			BaseURL:      "https://dashscope.aliyuncs.com/compatible-mode/v1",
			DefaultModel: "qwen-turbo",
		},
		{
			ID:           "zhipu",
			Name:         "智谱AI",
			Description:  "智谱AI GLM系列模型",
			BaseURL:      "https://open.bigmodel.cn/api/paas/v4/",
			DefaultModel: "glm-4-flash",
		},
		{
			ID:           "gemini",
			Name:         "Google Gemini",
			Description:  "Google Gemini模型（通过OpenAI兼容接口）",
			BaseURL:      "https://generativelanguage.googleapis.com/v1beta/openai",
			DefaultModel: "gemini-2.0-flash",
		},
		{
			ID:           "custom",
			Name:         "自定义",
			Description:  "自定义OpenAI兼容API（如one-api、new-api代理）",
			BaseURL:      "",
			DefaultModel: "",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    providers,
	})
}

// ========== AI服务状态 ==========

// AIServiceStatusResponse AI服务状态响应
type AIServiceStatusResponse struct {
	Provider  string `json:"provider"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Available bool   `json:"available"`
	Model     string `json:"model,omitempty"`
	BaseURL   string `json:"base_url,omitempty"`
	IsPrimary bool   `json:"is_primary"`
	LastError string `json:"last_error,omitempty"`
}

// getAIServicesStatus 获取所有AI服务状态
func (h *ConfigHandler) getAIServicesStatus(c *gin.Context) {
	var services []AIServiceStatusResponse

	// 获取用户选择的首选服务
	primaryProvider := h.App.Config.PrimaryAIService

	// 1. OpenAI兼容API
	openaiConfig := h.App.Config.OpenAICompatibleConfig
	openaiEnabled := openaiConfig != nil && openaiConfig.Enabled && openaiConfig.ApiKey != ""
	openaiService := AIServiceStatusResponse{
		Provider:  "openai_compatible",
		Name:      "自定义API",
		Enabled:   openaiEnabled,
		Available: openaiEnabled,
		IsPrimary: primaryProvider == "openai_compatible" && openaiEnabled,
	}
	if openaiConfig != nil {
		openaiService.Model = openaiConfig.Model
		openaiService.BaseURL = openaiConfig.BaseURL
		// 根据Provider显示更友好的名称
		switch openaiConfig.Provider {
		case "openai":
			openaiService.Name = "OpenAI"
		case "deepseek":
			openaiService.Name = "DeepSeek (兼容模式)"
		case "qwen":
			openaiService.Name = "通义千问"
		case "zhipu":
			openaiService.Name = "智谱AI"
		case "gemini":
			openaiService.Name = "Gemini (代理)"
		case "custom":
			openaiService.Name = "自定义API"
		}
	}
	services = append(services, openaiService)

	// 2. DeepSeek
	deepseekConfig := h.App.Config.DeepSeekTransConfig
	deepseekEnabled := deepseekConfig != nil && deepseekConfig.Enabled && deepseekConfig.ApiKey != ""
	deepseekService := AIServiceStatusResponse{
		Provider:  "deepseek",
		Name:      "DeepSeek",
		Enabled:   deepseekEnabled,
		Available: deepseekEnabled,
		IsPrimary: primaryProvider == "deepseek" && deepseekEnabled,
	}
	if deepseekConfig != nil {
		deepseekService.Model = deepseekConfig.Model
		deepseekService.BaseURL = deepseekConfig.Endpoint
	}
	services = append(services, deepseekService)

	// 3. Gemini（原生）
	geminiConfig := h.App.Config.GeminiConfig
	geminiEnabled := geminiConfig != nil && geminiConfig.Enabled && (geminiConfig.ApiKey != "" || len(geminiConfig.ApiKeys) > 0)
	geminiService := AIServiceStatusResponse{
		Provider:  "gemini",
		Name:      "Gemini（原生多模态）",
		Enabled:   geminiEnabled,
		Available: geminiEnabled,
		IsPrimary: primaryProvider == "gemini" && geminiEnabled,
	}
	if geminiConfig != nil {
		geminiService.Model = geminiConfig.Model
	}
	services = append(services, geminiService)

	// 如果没有设置首选或首选服务未启用，自动选择第一个启用的服务
	hasPrimary := false
	for _, svc := range services {
		if svc.IsPrimary {
			hasPrimary = true
			break
		}
	}
	if !hasPrimary {
		// 自动选择第一个启用的服务
		for i := range services {
			if services[i].Enabled {
				services[i].IsPrimary = true
				primaryProvider = services[i].Provider
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"services":         services,
			"primary_provider": primaryProvider,
			"has_available":    openaiEnabled || deepseekEnabled || geminiEnabled,
		},
	})
}

// SetPrimaryAIServiceRequest 设置首选AI服务请求
type SetPrimaryAIServiceRequest struct {
	Provider string `json:"provider"` // openai_compatible, deepseek, gemini
}

// setPrimaryAIService 设置首选AI服务
func (h *ConfigHandler) setPrimaryAIService(c *gin.Context) {
	var req SetPrimaryAIServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 验证提供商是否有效
	validProviders := map[string]bool{
		"openai_compatible": true,
		"deepseek":          true,
		"gemini":            true,
	}
	if !validProviders[req.Provider] {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid provider. Must be one of: openai_compatible, deepseek, gemini",
		})
		return
	}

	// 检查该服务是否已启用
	isEnabled := false
	switch req.Provider {
	case "openai_compatible":
		cfg := h.App.Config.OpenAICompatibleConfig
		isEnabled = cfg != nil && cfg.Enabled && cfg.ApiKey != ""
	case "deepseek":
		cfg := h.App.Config.DeepSeekTransConfig
		isEnabled = cfg != nil && cfg.Enabled && cfg.ApiKey != ""
	case "gemini":
		cfg := h.App.Config.GeminiConfig
		isEnabled = cfg != nil && cfg.Enabled && (cfg.ApiKey != "" || len(cfg.ApiKeys) > 0)
	}

	if !isEnabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "该AI服务未启用或未配置，请先配置并启用该服务",
		})
		return
	}

	// 更新配置
	h.App.Config.PrimaryAIService = req.Provider

	// 保存到配置文件
	if err := types.SaveConfig(h.App.Config); err != nil {
		h.App.Logger.Errorf("Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to save configuration: " + err.Error(),
		})
		return
	}

	h.App.Logger.Infof("✅ 首选AI服务已更新为: %s", req.Provider)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "首选AI服务设置成功",
		"data": gin.H{
			"primary_provider": req.Provider,
		},
	})
}

// ========== Gemini 原生配置（用于元数据生成） ==========

// GeminiConfigRequest Gemini配置请求
type GeminiConfigRequest struct {
	Enabled           *bool    `json:"enabled,omitempty"`
	ApiKey            *string  `json:"api_key,omitempty"`        // 单个 API Key（兼容旧配置）
	ApiKeys           []string `json:"api_keys,omitempty"`       // 多个 API Key（用于轮询）
	ClearApiKeys      *bool    `json:"clear_api_keys,omitempty"` // 是否清空所有 API Keys
	Model             *string  `json:"model,omitempty"`
	Timeout           *int     `json:"timeout,omitempty"`
	MaxTokens         *int     `json:"max_tokens,omitempty"`
	UseForMetadata    *bool    `json:"use_for_metadata,omitempty"`
	AnalyzeVideo      *bool    `json:"analyze_video,omitempty"`
	VideoSampleFrames *int     `json:"video_sample_frames,omitempty"`
}

// GeminiConfigResponse Gemini配置响应
type GeminiConfigResponse struct {
	Enabled           bool     `json:"enabled"`
	ApiKey            string   `json:"api_key"`        // 脱敏显示（兼容旧配置）
	ApiKeys           []string `json:"api_keys"`       // 多个 API Key（脱敏显示）
	ApiKeysCount      int      `json:"api_keys_count"` // API Key 数量
	Model             string   `json:"model"`
	Timeout           int      `json:"timeout"`
	MaxTokens         int      `json:"max_tokens"`
	UseForMetadata    bool     `json:"use_for_metadata"`
	AnalyzeVideo      bool     `json:"analyze_video"`
	VideoSampleFrames int      `json:"video_sample_frames"`
}

// getGeminiConfig 获取Gemini配置
func (h *ConfigHandler) getGeminiConfig(c *gin.Context) {
	config := h.App.Config.GeminiConfig
	if config == nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "success",
			"data": GeminiConfigResponse{
				Enabled:           false,
				ApiKey:            "",
				ApiKeys:           []string{},
				ApiKeysCount:      0,
				Model:             "gemini-2.0-flash",
				Timeout:           120,
				MaxTokens:         8000,
				UseForMetadata:    true,
				AnalyzeVideo:      true,
				VideoSampleFrames: 0,
			},
		})
		return
	}

	// 脱敏显示单个 API Key
	apiKeyMasked := ""
	if config.ApiKey != "" {
		if len(config.ApiKey) > 10 {
			apiKeyMasked = config.ApiKey[:6] + "..." + config.ApiKey[len(config.ApiKey)-4:]
		} else {
			apiKeyMasked = "***"
		}
	}

	// 脱敏显示多个 API Keys
	apiKeysMasked := make([]string, len(config.ApiKeys))
	for i, key := range config.ApiKeys {
		if len(key) > 10 {
			apiKeysMasked[i] = key[:6] + "..." + key[len(key)-4:]
		} else {
			apiKeysMasked[i] = "***"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": GeminiConfigResponse{
			Enabled:           config.Enabled,
			ApiKey:            apiKeyMasked,
			ApiKeys:           apiKeysMasked,
			ApiKeysCount:      config.GetApiKeysCount(),
			Model:             config.Model,
			Timeout:           config.Timeout,
			MaxTokens:         config.MaxTokens,
			UseForMetadata:    config.UseForMetadata,
			AnalyzeVideo:      config.AnalyzeVideo,
			VideoSampleFrames: config.VideoSampleFrames,
		},
	})
}

// updateGeminiConfig 更新Gemini配置
func (h *ConfigHandler) updateGeminiConfig(c *gin.Context) {
	var req GeminiConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// 确保配置对象存在
	if h.App.Config.GeminiConfig == nil {
		h.App.Config.GeminiConfig = &types.GeminiConfig{
			Enabled:           false,
			ApiKey:            "",
			ApiKeys:           []string{},
			Model:             "gemini-2.0-flash",
			Timeout:           120,
			MaxTokens:         8000,
			UseForMetadata:    true,
			AnalyzeVideo:      true,
			VideoSampleFrames: 0,
		}
	}

	config := h.App.Config.GeminiConfig

	// 更新配置字段
	if req.Enabled != nil {
		config.Enabled = *req.Enabled
	}
	if req.ApiKey != nil {
		config.ApiKey = *req.ApiKey
	}
	// 处理清空 API Keys 的请求
	if req.ClearApiKeys != nil && *req.ClearApiKeys {
		config.ApiKeys = []string{}
		config.ApiKey = ""
		h.App.Logger.Info("🗑️ Clearing all Gemini API Keys")
	} else if len(req.ApiKeys) > 0 {
		config.ApiKeys = req.ApiKeys
	}
	if req.Model != nil {
		config.Model = *req.Model
	}
	if req.Timeout != nil {
		config.Timeout = *req.Timeout
	}
	if req.MaxTokens != nil {
		config.MaxTokens = *req.MaxTokens
	}
	if req.UseForMetadata != nil {
		config.UseForMetadata = *req.UseForMetadata
	}
	if req.AnalyzeVideo != nil {
		config.AnalyzeVideo = *req.AnalyzeVideo
	}
	if req.VideoSampleFrames != nil {
		config.VideoSampleFrames = *req.VideoSampleFrames
	}

	// 保存配置到文件
	if err := types.SaveConfig(h.App.Config); err != nil {
		h.App.Logger.Errorf("Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to save configuration: " + err.Error(),
		})
		return
	}

	h.App.Config.GeminiConfig = config
	h.App.Logger.Infof("✅ Gemini configuration updated successfully (API Keys: %d)", config.GetApiKeysCount())

	// 脱敏显示多个 API Keys
	apiKeysMasked := make([]string, len(config.ApiKeys))
	for i, key := range config.ApiKeys {
		if len(key) > 10 {
			apiKeysMasked[i] = key[:6] + "..." + key[len(key)-4:]
		} else {
			apiKeysMasked[i] = "***"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Configuration updated successfully",
		"data": GeminiConfigResponse{
			Enabled:           config.Enabled,
			ApiKey:            maskApiKey(config.ApiKey),
			ApiKeys:           apiKeysMasked,
			ApiKeysCount:      config.GetApiKeysCount(),
			Model:             config.Model,
			Timeout:           config.Timeout,
			MaxTokens:         config.MaxTokens,
			UseForMetadata:    config.UseForMetadata,
			AnalyzeVideo:      config.AnalyzeVideo,
			VideoSampleFrames: config.VideoSampleFrames,
		},
	})
}

// ========== Gemini API Key 验证 ==========

// ApiKeyValidationResult 单个 API Key 的验证结果
type ApiKeyValidationResult struct {
	Key     string `json:"key"`     // 脱敏后的 Key
	Index   int    `json:"index"`   // Key 的索引
	Valid   bool   `json:"valid"`   // 是否有效
	Message string `json:"message"` // 验证消息
}

// ValidateGeminiApiKeysResponse 验证响应
type ValidateGeminiApiKeysResponse struct {
	TotalKeys   int                      `json:"total_keys"`   // 总 Key 数量
	ValidKeys   int                      `json:"valid_keys"`   // 有效 Key 数量
	InvalidKeys int                      `json:"invalid_keys"` // 无效 Key 数量
	Results     []ApiKeyValidationResult `json:"results"`      // 每个 Key 的验证结果
	AutoRemoved int                      `json:"auto_removed"` // 自动移除的无效 Key 数量
}

// validateGeminiApiKeys 验证所有 Gemini API Keys
func (h *ConfigHandler) validateGeminiApiKeys(c *gin.Context) {
	config := h.App.Config.GeminiConfig
	if config == nil || len(config.ApiKeys) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "No API Keys configured",
			"data": ValidateGeminiApiKeysResponse{
				TotalKeys:   0,
				ValidKeys:   0,
				InvalidKeys: 0,
				Results:     []ApiKeyValidationResult{},
			},
		})
		return
	}

	h.App.Logger.Info("🔍 开始验证 Gemini API Keys...")

	// 并发验证所有 API Keys
	var wg sync.WaitGroup
	results := make([]ApiKeyValidationResult, len(config.ApiKeys))

	for i, apiKey := range config.ApiKeys {
		wg.Add(1)
		go func(index int, key string) {
			defer wg.Done()

			result := ApiKeyValidationResult{
				Key:   maskApiKey(key),
				Index: index,
			}

			// 创建临时客户端测试连接
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			client, err := genai.NewClient(ctx, option.WithAPIKey(key))
			if err != nil {
				result.Valid = false
				result.Message = fmt.Sprintf("创建客户端失败: %v", err)
				results[index] = result
				return
			}
			defer client.Close()

			// 测试 API 调用
			model := client.GenerativeModel(config.Model)
			model.SetMaxOutputTokens(10)
			model.SetTemperature(0.1)

			_, err = model.GenerateContent(ctx, genai.Text("Hi"))
			if err != nil {
				result.Valid = false
				errMsg := err.Error()
				if strings.Contains(errMsg, "leaked") {
					result.Message = "⚠️ API Key 已泄露，请更换"
				} else if strings.Contains(errMsg, "invalid") || strings.Contains(errMsg, "API key not valid") {
					result.Message = "❌ API Key 无效"
				} else if strings.Contains(errMsg, "quota") {
					result.Message = "⚠️ 配额已用尽"
				} else if strings.Contains(errMsg, "403") {
					result.Message = "❌ 访问被拒绝: " + errMsg
				} else {
					result.Message = "❌ 验证失败: " + errMsg
				}
			} else {
				result.Valid = true
				result.Message = "✅ 有效"
			}

			results[index] = result
		}(i, apiKey)
	}

	wg.Wait()

	// 统计结果
	validCount := 0
	invalidCount := 0
	for _, r := range results {
		if r.Valid {
			validCount++
		} else {
			invalidCount++
		}
	}

	h.App.Logger.Infof("🔍 Gemini API Keys 验证完成: %d 有效, %d 无效", validCount, invalidCount)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": fmt.Sprintf("验证完成: %d 有效, %d 无效", validCount, invalidCount),
		"data": ValidateGeminiApiKeysResponse{
			TotalKeys:   len(config.ApiKeys),
			ValidKeys:   validCount,
			InvalidKeys: invalidCount,
			Results:     results,
			AutoRemoved: 0,
		},
	})
}

// getGeminiModels 获取 Gemini 可用模型列表
func (h *ConfigHandler) getGeminiModels(c *gin.Context) {
	config := h.App.Config.GeminiConfig
	if config == nil || len(config.ApiKeys) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "No API Keys configured",
		})
		return
	}

	// 使用第一个 API Key 获取模型列表
	apiKey := config.ApiKeys[0]

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建客户端失败: " + err.Error(),
		})
		return
	}
	defer client.Close()

	// 获取模型列表
	iter := client.ListModels(ctx)
	var models []string
	for {
		m, err := iter.Next()
		if err != nil {
			break
		}
		// 只返回支持生成内容的模型
		if strings.Contains(m.Name, "gemini") {
			// 提取模型名称（去掉 "models/" 前缀）
			modelName := strings.TrimPrefix(m.Name, "models/")
			models = append(models, modelName)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"models": models,
		},
	})
}
