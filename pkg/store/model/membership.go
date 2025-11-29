package model

import (
	"time"

	"gorm.io/gorm"
)

// UserMembership 用户会员信息
type UserMembership struct {
	gorm.Model
	UserID     uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	Tier       string     `gorm:"type:varchar(20);default:'free'" json:"tier"`     // free, basic, pro, enterprise
	Status     string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, expired, cancelled
	StartDate  time.Time  `json:"start_date"`
	ExpireDate *time.Time `json:"expire_date,omitempty"`
	AutoRenew  bool       `gorm:"default:false" json:"auto_renew"`
}

// MembershipOrder 会员订单
type MembershipOrder struct {
	gorm.Model
	OrderNo       string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	PlanID        string     `gorm:"type:varchar(20);not null" json:"plan_id"` // free, basic, pro, enterprise
	Amount        float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Currency      string     `gorm:"type:varchar(10);default:'CNY'" json:"currency"`
	Status        string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, paid, failed, refunded, cancelled
	PaymentMethod string     `gorm:"type:varchar(20)" json:"payment_method"`           // alipay, wechat, stripe
	PaymentID     string     `gorm:"type:varchar(128)" json:"payment_id,omitempty"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ExpireAt      *time.Time `json:"expire_at,omitempty"`
	Months        int        `gorm:"default:1" json:"months"` // 购买月数
	Remark        string     `gorm:"type:text" json:"remark,omitempty"`
}

// License 设备授权
type License struct {
	gorm.Model
	LicenseKey  string     `gorm:"type:varchar(128);uniqueIndex;not null" json:"license_key"`
	UserID      uint       `gorm:"index" json:"user_id"`
	DeviceID    string     `gorm:"type:varchar(128);index" json:"device_id"`
	DeviceName  string     `gorm:"type:varchar(128)" json:"device_name"`
	Tier        string     `gorm:"type:varchar(20);default:'free'" json:"tier"`
	Status      string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, revoked, expired
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

// UsageRecord 使用记录（用于统计每日限额）
type UsageRecord struct {
	gorm.Model
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Date      string    `gorm:"type:varchar(10);index;not null" json:"date"` // YYYY-MM-DD
	Feature   string    `gorm:"type:varchar(50);not null" json:"feature"`    // video_process, ai_translation, etc.
	Count     int       `gorm:"default:0" json:"count"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (UserMembership) TableName() string {
	return "cw_user_memberships"
}

func (MembershipOrder) TableName() string {
	return "cw_membership_orders"
}

func (License) TableName() string {
	return "cw_licenses"
}

func (UsageRecord) TableName() string {
	return "cw_usage_records"
}
