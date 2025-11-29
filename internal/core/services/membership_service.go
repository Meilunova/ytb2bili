package services

import (
	"errors"
	"time"

	"github.com/difyz9/ytb2bili/internal/core/types"
	"github.com/difyz9/ytb2bili/pkg/store/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MembershipService 会员服务
type MembershipService struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

// NewMembershipService 创建会员服务
func NewMembershipService(db *gorm.DB, logger *zap.SugaredLogger) *MembershipService {
	return &MembershipService{
		DB:     db,
		Logger: logger,
	}
}

// GetUserMembership 获取用户会员信息
func (s *MembershipService) GetUserMembership(userID uint) (*model.UserMembership, error) {
	var membership model.UserMembership
	err := s.DB.Where("user_id = ?", userID).First(&membership).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 返回默认免费会员
			return &model.UserMembership{
				UserID:    userID,
				Tier:      string(types.TierFree),
				Status:    "active",
				StartDate: time.Now(),
			}, nil
		}
		return nil, err
	}
	return &membership, nil
}

// GetMembershipTier 获取用户会员等级
func (s *MembershipService) GetMembershipTier(userID uint) types.MembershipTier {
	membership, err := s.GetUserMembership(userID)
	if err != nil {
		return types.TierFree
	}

	// 检查是否过期
	if membership.ExpireDate != nil && membership.ExpireDate.Before(time.Now()) {
		return types.TierFree
	}

	return types.MembershipTier(membership.Tier)
}

// GetMembershipLimits 获取用户功能限制
func (s *MembershipService) GetMembershipLimits(userID uint) types.MembershipLimits {
	tier := s.GetMembershipTier(userID)
	return types.GetMembershipLimits(tier)
}

// CanUseFeature 检查用户是否可以使用某功能
func (s *MembershipService) CanUseFeature(userID uint, feature string) bool {
	tier := s.GetMembershipTier(userID)
	return types.CanUseFeature(tier, feature)
}

// CheckDailyLimit 检查每日视频处理限制
func (s *MembershipService) CheckDailyLimit(userID uint) (bool, int, int) {
	limits := s.GetMembershipLimits(userID)

	// 无限制
	if limits.VideosPerDay == -1 {
		return true, -1, 0
	}

	// 统计今日已处理的视频数量
	today := time.Now().Format("2006-01-02")
	var count int64
	s.DB.Model(&model.SavedVideo{}).
		Where("user_id = ? AND DATE(created_at) = ?", userID, today).
		Count(&count)

	remaining := limits.VideosPerDay - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return int(count) < limits.VideosPerDay, limits.VideosPerDay, int(count)
}

// CheckBatchLimit 检查批量提交限制
func (s *MembershipService) CheckBatchLimit(userID uint, batchSize int) bool {
	limits := s.GetMembershipLimits(userID)
	return batchSize <= limits.BatchSize
}

// UpgradeMembership 升级会员
func (s *MembershipService) UpgradeMembership(userID uint, tier types.MembershipTier, months int) error {
	membership, err := s.GetUserMembership(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now()
	expireDate := now.AddDate(0, months, 0)

	// 如果已有会员且未过期，在原有基础上延长
	if membership.ID != 0 && membership.ExpireDate != nil && membership.ExpireDate.After(now) {
		expireDate = membership.ExpireDate.AddDate(0, months, 0)
	}

	newMembership := model.UserMembership{
		UserID:     userID,
		Tier:       string(tier),
		Status:     "active",
		StartDate:  now,
		ExpireDate: &expireDate,
	}

	if membership.ID != 0 {
		newMembership.ID = membership.ID
		return s.DB.Save(&newMembership).Error
	}

	return s.DB.Create(&newMembership).Error
}

// GetAllPlans 获取所有会员计划
func (s *MembershipService) GetAllPlans() []types.MembershipPlan {
	plans := make([]types.MembershipPlan, 0, len(types.MembershipPlans))
	for _, plan := range types.MembershipPlans {
		plans = append(plans, plan)
	}
	return plans
}
