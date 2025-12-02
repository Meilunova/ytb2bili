package membership

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MembershipMiddleware 会员中间件
type MembershipMiddleware struct {
	store   MembershipStore
	checker *FeatureChecker
	quota   *QuotaService
}

// NewMembershipMiddleware 创建会员中间件
func NewMembershipMiddleware(store MembershipStore) *MembershipMiddleware {
	checker := NewFeatureChecker(store)
	return &MembershipMiddleware{
		store:   store,
		checker: checker,
		quota:   NewQuotaService(store, checker),
	}
}

// ContextKey 上下文键类型
type ContextKey string

const (
	// ContextKeyUserID 用户ID上下文键
	ContextKeyUserID ContextKey = "user_id"
	// ContextKeyMembership 会员信息上下文键
	ContextKeyMembership ContextKey = "membership"
	// ContextKeyTier 会员等级上下文键
	ContextKeyTier ContextKey = "membership_tier"
)

// RequireFeature 检查功能权限的中间件
func (m *MembershipMiddleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		result := m.checker.CanUseFeature(c.Request.Context(), userID, feature)
		if !result.Allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"code":       403,
				"message":    result.Reason,
				"feature":    feature,
				"suggestion": result.Upgrade,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireQuota 检查配额的中间件
func (m *MembershipMiddleware) RequireQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		hasQuota := m.quota.HasQuota(c.Request.Context(), userID)
		if !hasQuota {
			quotaInfo, _ := m.quota.GetQuotaInfo(c.Request.Context(), userID)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "今日配额已用完",
				"quota":   quotaInfo,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireBatch 检查批量操作限制的中间件
func (m *MembershipMiddleware) RequireBatch(countGetter func(*gin.Context) int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		count := countGetter(c)
		result := m.checker.CanBatchSubmit(c.Request.Context(), userID, count)
		if !result.Allowed {
			limit := m.quota.GetBatchLimit(c.Request.Context(), userID)
			c.JSON(http.StatusForbidden, gin.H{
				"code":       403,
				"message":    result.Reason,
				"requested":  count,
				"max_limit":  limit,
				"suggestion": result.Upgrade,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTier 检查会员等级的中间件
func (m *MembershipMiddleware) RequireTier(minTier Tier) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			c.Abort()
			return
		}

		tier := m.checker.GetUserTier(c.Request.Context(), userID)

		if CompareTiers(tier, minTier) < 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"code":         403,
				"message":      "会员等级不足",
				"current_tier": string(tier),
				"require_tier": string(minTier),
			})
			c.Abort()
			return
		}

		// 将等级信息存入上下文
		c.Set(string(ContextKeyTier), tier)
		c.Next()
	}
}

// InjectMembership 注入会员信息到上下文的中间件
func (m *MembershipMiddleware) InjectMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := m.getUserID(c)
		if userID == "" {
			c.Next()
			return
		}

		// 获取会员信息
		membership, err := m.checker.store.GetUserMembership(c.Request.Context(), userID)
		if err == nil {
			c.Set(string(ContextKeyMembership), membership)
			c.Set(string(ContextKeyTier), membership.GetEffectiveTier())
		}

		c.Next()
	}
}

// ConsumeQuota 消耗配额的中间件（在请求成功后调用）
func (m *MembershipMiddleware) ConsumeQuota() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只有请求成功才消耗配额
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			userID := m.getUserID(c)
			if userID != "" {
				// 异步消耗配额，不阻塞响应
				go func() {
					m.quota.ConsumeQuota(c.Request.Context(), userID)
				}()
			}
		}
	}
}

// getUserID 从上下文获取用户ID
func (m *MembershipMiddleware) getUserID(c *gin.Context) string {
	// 优先从 JWT 认证中间件设置的 context 获取
	if userID, exists := c.Get("user_id"); exists {
		switch v := userID.(type) {
		case string:
			return v
		case uint:
			return strconv.FormatUint(uint64(v), 10)
		case int:
			return strconv.Itoa(v)
		}
	}

	// 兼容旧的 context key
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

	// 尝试从 Header 获取 (用于测试/调试)
	if userID := c.GetHeader("X-User-ID"); userID != "" {
		return userID
	}

	return ""
}

// GetMembershipFromContext 从上下文获取会员信息
func GetMembershipFromContext(c *gin.Context) (*UserMembership, bool) {
	if m, exists := c.Get(string(ContextKeyMembership)); exists {
		if membership, ok := m.(*UserMembership); ok {
			return membership, true
		}
	}
	return nil, false
}

// GetTierFromContext 从上下文获取会员等级
func GetTierFromContext(c *gin.Context) (Tier, bool) {
	if t, exists := c.Get(string(ContextKeyTier)); exists {
		if tier, ok := t.(Tier); ok {
			return tier, true
		}
	}
	return TierFree, false
}
