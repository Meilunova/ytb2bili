package membership

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MembershipHandler 会员 API Handler
type MembershipHandler struct {
	store     MembershipStore
	checker   *FeatureChecker
	quota     *QuotaService
	boostPack *BoostPackService
}

// NewMembershipHandler 创建会员 Handler
func NewMembershipHandler(store MembershipStore) *MembershipHandler {
	checker := NewFeatureChecker(store)
	return &MembershipHandler{
		store:     store,
		checker:   checker,
		quota:     NewQuotaService(store, checker),
		boostPack: NewBoostPackService(store),
	}
}

// RegisterRoutes 注册会员相关路由
func (h *MembershipHandler) RegisterRoutes(router *gin.RouterGroup) {
	membership := router.Group("/membership")
	{
		// 会员信息
		membership.GET("/info", h.GetMembershipInfo)
		membership.GET("/tiers", h.GetAllTiers)

		// 配额相关
		membership.GET("/quota", h.GetQuotaInfo)

		// 功能检查
		membership.GET("/features", h.GetAvailableFeatures)
		membership.GET("/features/:feature/check", h.CheckFeature)

		// 加油包
		membership.GET("/boost-pack", h.GetBoostPackStatus)
		membership.POST("/boost-pack/purchase", h.PurchaseBoostPack)

		// 管理接口（需要管理员权限）
		membership.PUT("/admin/user/:userId/tier", h.AdminSetUserTier)
		membership.POST("/admin/user/:userId/boost-pack", h.AdminAddBoostPack)
	}
}

// ========== 响应结构体 ==========

// MembershipInfoResponse 会员信息响应
type MembershipInfoResponse struct {
	UserID         string     `json:"user_id"`
	Tier           string     `json:"tier"`
	TierName       string     `json:"tier_name"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	DaysRemaining  int        `json:"days_remaining"`
	IsExpired      bool       `json:"is_expired"`
	DailyLimit     int        `json:"daily_limit"`
	BatchLimit     int        `json:"batch_limit"`
	Priority       int        `json:"priority"`
	SubscriptionID string     `json:"subscription_id,omitempty"`
}

// TierInfoResponse 等级信息响应
type TierInfoResponse struct {
	Tier        string  `json:"tier"`
	Name        string  `json:"name"`
	DailyLimit  int     `json:"daily_limit"`
	BatchLimit  int     `json:"batch_limit"`
	Priority    int     `json:"priority"`
	Price       float64 `json:"price,omitempty"`
	Description string  `json:"description,omitempty"`
}

// FeatureCheckResponse 功能检查响应
type FeatureCheckResponse struct {
	Feature    string `json:"feature"`
	Allowed    bool   `json:"allowed"`
	Reason     string `json:"reason,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
}

// BoostPackStatusResponse 加油包状态响应
type BoostPackStatusResponse struct {
	HasPack         bool       `json:"has_pack"`
	VideosRemaining int        `json:"videos_remaining"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	DaysRemaining   int        `json:"days_remaining"`
}

// ========== Handler 方法 ==========

// GetMembershipInfo 获取当前用户会员信息
func (h *MembershipHandler) GetMembershipInfo(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	membership, err := h.store.GetUserMembership(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取会员信息失败: " + err.Error(),
		})
		return
	}

	tier := membership.GetEffectiveTier()
	config := GetTierConfig(tier)

	var expiresAt *time.Time
	if !membership.ExpiresAt.IsZero() {
		expiresAt = &membership.ExpiresAt
	}

	response := MembershipInfoResponse{
		UserID:         userID,
		Tier:           string(tier),
		TierName:       config.Name,
		ExpiresAt:      expiresAt,
		DaysRemaining:  membership.DaysUntilExpiry(),
		IsExpired:      membership.IsExpired(),
		DailyLimit:     config.Limits.VideosPerDay,
		BatchLimit:     config.Limits.BatchSize,
		Priority:       config.Priority,
		SubscriptionID: membership.SubscriptionID,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// GetAllTiers 获取所有会员等级信息
func (h *MembershipHandler) GetAllTiers(c *gin.Context) {
	configs := GetAllTierConfigs()
	tiers := make([]TierInfoResponse, 0, len(configs))

	descriptions := map[Tier]string{
		TierFree:       "免费体验，每日5个视频",
		TierBasic:      "基础会员，每日20个视频，支持AI翻译",
		TierPro:        "专业会员，每日100个视频，支持所有AI功能",
		TierEnterprise: "企业会员，无限制使用，专属支持",
	}

	for _, config := range configs {
		tiers = append(tiers, TierInfoResponse{
			Tier:        string(config.Tier),
			Name:        config.Name,
			DailyLimit:  config.Limits.VideosPerDay,
			BatchLimit:  config.Limits.BatchSize,
			Priority:    config.Priority,
			Price:       config.Price,
			Description: descriptions[config.Tier],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    tiers,
	})
}

// GetQuotaInfo 获取配额信息
func (h *MembershipHandler) GetQuotaInfo(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	quotaInfo, err := h.quota.GetQuotaInfo(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取配额信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    quotaInfo,
	})
}

// GetAvailableFeatures 获取当前用户可用功能列表
func (h *MembershipHandler) GetAvailableFeatures(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	tier := h.checker.GetUserTier(c.Request.Context(), userID)
	config := GetTierConfig(tier)

	// 构建功能列表
	features := []string{}
	if config.Features.AITranslation {
		features = append(features, "ai_translation")
	}
	if config.Features.TranslationOptimize {
		features = append(features, "translation_optimize")
	}
	if config.Features.AITitleGeneration {
		features = append(features, "ai_title_generation")
	}
	if config.Features.GeminiVideoAnalysis {
		features = append(features, "gemini_video_analysis")
	}
	if config.Features.AutoUpload {
		features = append(features, "auto_upload")
	}
	if config.Features.PriorityQueue {
		features = append(features, "priority_queue")
	}
	if config.Features.APIAccess {
		features = append(features, "api_access")
	}
	if config.Features.CustomTemplate {
		features = append(features, "custom_template")
	}
	if config.Features.DataExport {
		features = append(features, "data_export")
	}
	if config.Features.TeamCollaboration {
		features = append(features, "team_collaboration")
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"tier":     string(tier),
			"features": features,
		},
	})
}

// CheckFeature 检查特定功能是否可用
func (h *MembershipHandler) CheckFeature(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	featureName := c.Param("feature")
	result := h.checker.CanUseFeature(c.Request.Context(), userID, featureName)

	response := FeatureCheckResponse{
		Feature:    featureName,
		Allowed:    result.Allowed,
		Reason:     result.Reason,
		Suggestion: result.Upgrade,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// GetBoostPackStatus 获取加油包状态
func (h *MembershipHandler) GetBoostPackStatus(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	status, err := h.boostPack.GetStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取加油包状态失败: " + err.Error(),
		})
		return
	}

	var expiresAt *time.Time
	if !status.ExpiresAt.IsZero() {
		expiresAt = &status.ExpiresAt
	}

	response := BoostPackStatusResponse{
		HasPack:         status.HasPack,
		VideosRemaining: status.VideosRemaining,
		ExpiresAt:       expiresAt,
		DaysRemaining:   status.DaysRemaining,
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// PurchaseBoostPack 购买加油包
func (h *MembershipHandler) PurchaseBoostPack(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	var req struct {
		PackType string `json:"pack_type" binding:"required"` // small, medium, large
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	packType := BoostPackType(req.PackType)
	config, ok := GetBoostPackConfig(packType)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的加油包类型",
		})
		return
	}

	// TODO: 这里应该集成支付系统
	// 目前直接添加加油包（仅用于测试）
	result, err := h.boostPack.Purchase(c.Request.Context(), userID, packType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "购买失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": result.Message,
		"data": gin.H{
			"pack_type":    string(packType),
			"videos_added": config.Videos,
			"total_videos": result.TotalVideos,
			"expires_at":   result.ExpiresAt,
		},
	})
}

// ========== 管理接口 ==========

// AdminSetUserTier 管理员设置用户会员等级
func (h *MembershipHandler) AdminSetUserTier(c *gin.Context) {
	// TODO: 添加管理员权限检查

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID不能为空",
		})
		return
	}

	var req struct {
		Tier      string `json:"tier" binding:"required"`
		ValidDays int    `json:"valid_days"` // 有效天数，0表示永久
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	tier := Tier(req.Tier)
	if !isValidTier(tier) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的会员等级",
		})
		return
	}

	var expiresAt time.Time
	if req.ValidDays > 0 {
		expiresAt = time.Now().AddDate(0, 0, req.ValidDays)
	}

	membership := &UserMembership{
		UserID:    targetUserID,
		Tier:      tier,
		ExpiresAt: expiresAt,
		UpdatedAt: time.Now(),
	}

	if err := h.store.SaveUserMembership(c.Request.Context(), membership); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "设置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "设置成功",
		"data": gin.H{
			"user_id":    targetUserID,
			"tier":       string(tier),
			"expires_at": expiresAt,
		},
	})
}

// AdminAddBoostPack 管理员为用户添加加油包
func (h *MembershipHandler) AdminAddBoostPack(c *gin.Context) {
	// TODO: 添加管理员权限检查

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户ID不能为空",
		})
		return
	}

	var req struct {
		Videos    int `json:"videos" binding:"required,min=1"`
		ValidDays int `json:"valid_days" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	pack := &UserBoostPack{
		UserID:          targetUserID,
		VideosRemaining: req.Videos,
		ExpiresAt:       time.Now().AddDate(0, 0, req.ValidDays),
	}

	if err := h.store.SaveUserBoostPack(c.Request.Context(), pack); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "添加失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "添加成功",
		"data": gin.H{
			"user_id":    targetUserID,
			"videos":     req.Videos,
			"valid_days": req.ValidDays,
		},
	})
}

// ========== 辅助方法 ==========

// getUserID 从上下文获取用户ID
func (h *MembershipHandler) getUserID(c *gin.Context) string {
	// 优先从上下文获取（由认证中间件设置）
	if userID, exists := c.Get(string(ContextKeyUserID)); exists {
		switch v := userID.(type) {
		case string:
			return v
		case uint:
			return strconv.FormatUint(uint64(v), 10)
		case int:
			return strconv.Itoa(v)
		}
	}

	// 尝试从 Header 获取
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}

	return ""
}

// isValidTier 检查是否是有效的会员等级
func isValidTier(tier Tier) bool {
	switch tier {
	case TierFree, TierBasic, TierPro, TierEnterprise:
		return true
	default:
		return false
	}
}
