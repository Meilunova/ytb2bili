package middleware

import (
	"net/http"

	"github.com/difyz9/ytb2bili/internal/core/services"
	"github.com/difyz9/ytb2bili/internal/core/types"
	"github.com/gin-gonic/gin"
)

// MembershipMiddleware 会员权限中间件
type MembershipMiddleware struct {
	MembershipService *services.MembershipService
}

// NewMembershipMiddleware 创建会员中间件
func NewMembershipMiddleware(membershipService *services.MembershipService) *MembershipMiddleware {
	return &MembershipMiddleware{
		MembershipService: membershipService,
	}
}

// RequireFeature 要求特定功能权限
func (m *MembershipMiddleware) RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		if !m.MembershipService.CanUseFeature(userID, feature) {
			tier := m.MembershipService.GetMembershipTier(userID)
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "当前会员等级不支持此功能，请升级会员",
				"data": gin.H{
					"current_tier":     tier,
					"required_feature": feature,
					"upgrade_url":      "/membership/upgrade",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTier 要求特定会员等级
func (m *MembershipMiddleware) RequireTier(minTier types.MembershipTier) gin.HandlerFunc {
	tierOrder := map[types.MembershipTier]int{
		types.TierFree:       0,
		types.TierBasic:      1,
		types.TierPro:        2,
		types.TierEnterprise: 3,
	}

	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		currentTier := m.MembershipService.GetMembershipTier(userID)
		if tierOrder[currentTier] < tierOrder[minTier] {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要更高的会员等级",
				"data": gin.H{
					"current_tier":  currentTier,
					"required_tier": minTier,
					"upgrade_url":   "/membership/upgrade",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CheckDailyLimit 检查每日限额
func (m *MembershipMiddleware) CheckDailyLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		canProcess, limit, used := m.MembershipService.CheckDailyLimit(userID)
		if !canProcess {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "今日视频处理次数已达上限",
				"data": gin.H{
					"daily_limit": limit,
					"used":        used,
					"remaining":   0,
					"upgrade_url": "/membership/upgrade",
				},
			})
			c.Abort()
			return
		}

		// 将限额信息存入上下文
		c.Set("daily_limit", limit)
		c.Set("daily_used", used)
		c.Next()
	}
}

// InjectMembershipInfo 注入会员信息到上下文
func (m *MembershipMiddleware) InjectMembershipInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		if userID == 0 {
			c.Next()
			return
		}

		tier := m.MembershipService.GetMembershipTier(userID)
		limits := m.MembershipService.GetMembershipLimits(userID)

		c.Set("membership_tier", tier)
		c.Set("membership_limits", limits)
		c.Next()
	}
}
