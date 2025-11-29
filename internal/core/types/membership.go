package types

// MembershipTier 会员等级
type MembershipTier string

const (
	TierFree       MembershipTier = "free"
	TierBasic      MembershipTier = "basic"
	TierPro        MembershipTier = "pro"
	TierEnterprise MembershipTier = "enterprise"
)

// MembershipLimits 会员功能限制
type MembershipLimits struct {
	VideosPerDay   int  `json:"videos_per_day"`  // 每日视频处理数量，-1 表示无限
	BatchSize      int  `json:"batch_size"`      // 批量提交大小
	AITranslation  bool `json:"ai_translation"`  // AI 字幕翻译
	AutoUpload     bool `json:"auto_upload"`     // 自动上传到B站
	PriorityQueue  bool `json:"priority_queue"`  // 优先处理队列
	APIAccess      bool `json:"api_access"`      // API 接口访问
	VideoAnalysis  bool `json:"video_analysis"`  // Gemini 视频分析
	CustomMetadata bool `json:"custom_metadata"` // 自定义元数据模板
}

// MembershipPlan 会员计划配置
type MembershipPlan struct {
	ID            MembershipTier   `json:"id"`
	Name          string           `json:"name"`
	Price         float64          `json:"price"`
	Period        string           `json:"period"` // month, year, lifetime
	OriginalPrice float64          `json:"original_price,omitempty"`
	Popular       bool             `json:"popular,omitempty"`
	Features      []string         `json:"features"`
	Limits        MembershipLimits `json:"limits"`
}

// 预定义会员计划
var MembershipPlans = map[MembershipTier]MembershipPlan{
	TierFree: {
		ID:     TierFree,
		Name:   "免费版",
		Price:  0,
		Period: "lifetime",
		Features: []string{
			"每日 5 个视频",
			"单个视频提交",
			"基础字幕下载",
			"社区支持",
		},
		Limits: MembershipLimits{
			VideosPerDay:   5,
			BatchSize:      1,
			AITranslation:  false,
			AutoUpload:     false,
			PriorityQueue:  false,
			APIAccess:      false,
			VideoAnalysis:  false,
			CustomMetadata: false,
		},
	},
	TierBasic: {
		ID:            TierBasic,
		Name:          "基础版",
		Price:         29,
		Period:        "month",
		OriginalPrice: 39,
		Features: []string{
			"每日 20 个视频",
			"批量提交（最多5个）",
			"AI 字幕翻译",
			"优先邮件支持",
			"无广告体验",
		},
		Limits: MembershipLimits{
			VideosPerDay:   20,
			BatchSize:      5,
			AITranslation:  true,
			AutoUpload:     false,
			PriorityQueue:  false,
			APIAccess:      false,
			VideoAnalysis:  false,
			CustomMetadata: false,
		},
	},
	TierPro: {
		ID:            TierPro,
		Name:          "专业版",
		Price:         99,
		Period:        "month",
		OriginalPrice: 129,
		Popular:       true,
		Features: []string{
			"每日 100 个视频",
			"批量提交（最多20个）",
			"AI 字幕翻译",
			"自动上传到B站",
			"优先处理队列",
			"专属客服支持",
			"高级数据分析",
		},
		Limits: MembershipLimits{
			VideosPerDay:   100,
			BatchSize:      20,
			AITranslation:  true,
			AutoUpload:     true,
			PriorityQueue:  true,
			APIAccess:      false,
			VideoAnalysis:  true,
			CustomMetadata: true,
		},
	},
	TierEnterprise: {
		ID:     TierEnterprise,
		Name:   "企业版",
		Price:  299,
		Period: "month",
		Features: []string{
			"无限视频处理",
			"批量提交（最多100个）",
			"AI 字幕翻译",
			"自动上传到B站",
			"最高优先级队列",
			"API 接口访问",
			"团队协作功能",
			"专属技术支持",
			"定制化服务",
		},
		Limits: MembershipLimits{
			VideosPerDay:   -1, // 无限
			BatchSize:      100,
			AITranslation:  true,
			AutoUpload:     true,
			PriorityQueue:  true,
			APIAccess:      true,
			VideoAnalysis:  true,
			CustomMetadata: true,
		},
	},
}

// GetMembershipLimits 获取指定等级的功能限制
func GetMembershipLimits(tier MembershipTier) MembershipLimits {
	if plan, ok := MembershipPlans[tier]; ok {
		return plan.Limits
	}
	return MembershipPlans[TierFree].Limits
}

// CanUseFeature 检查指定等级是否可以使用某功能
func CanUseFeature(tier MembershipTier, feature string) bool {
	limits := GetMembershipLimits(tier)
	switch feature {
	case "ai_translation":
		return limits.AITranslation
	case "auto_upload":
		return limits.AutoUpload
	case "priority_queue":
		return limits.PriorityQueue
	case "api_access":
		return limits.APIAccess
	case "video_analysis":
		return limits.VideoAnalysis
	case "custom_metadata":
		return limits.CustomMetadata
	default:
		return true
	}
}
