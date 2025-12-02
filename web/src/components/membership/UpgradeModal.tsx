'use client';

import { useState, useEffect } from 'react';
import { membershipApi } from '@/lib/api';
import type { TierConfig, MembershipTier } from '@/types';

interface UpgradeModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentTier?: MembershipTier;
}

const TIER_FEATURES: Record<MembershipTier, string[]> = {
  free: ['每日 5 个视频', '基础功能'],
  basic: ['每日 20 个视频', 'AI 字幕翻译', 'AI 标题生成', '自定义模板', '批量处理 5 个'],
  pro: ['每日 100 个视频', '所有基础版功能', '翻译质量优化', 'Gemini 视频分析', '自动上传', '优先队列', '数据导出', '批量处理 20 个'],
  enterprise: ['无限视频', '所有专业版功能', 'API 访问', '团队协作', '专属支持', '批量处理 100 个'],
};

const TIER_ICONS: Record<MembershipTier, string> = {
  free: '🆓',
  basic: '⭐',
  pro: '💎',
  enterprise: '👑',
};

export default function UpgradeModal({ isOpen, onClose, currentTier = 'free' }: UpgradeModalProps) {
  const [tiers, setTiers] = useState<TierConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedTier, setSelectedTier] = useState<MembershipTier | null>(null);
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly');

  useEffect(() => {
    if (isOpen) {
      fetchTiers();
    }
  }, [isOpen]);

  const fetchTiers = async () => {
    try {
      setLoading(true);
      const res = await membershipApi.getAllTiers();
      if (res.code === 0) {
        setTiers(res.data);
      }
    } catch (err) {
      console.error('获取等级信息失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleUpgrade = (tier: MembershipTier) => {
    setSelectedTier(tier);
    // TODO: 集成支付系统
    alert(`升级到 ${tier} 功能即将上线，敬请期待！`);
  };

  if (!isOpen) return null;

  const tierOrder: MembershipTier[] = ['free', 'basic', 'pro', 'enterprise'];
  const currentTierIndex = tierOrder.indexOf(currentTier);

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      {/* 背景遮罩 */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
        onClick={onClose}
      ></div>

      {/* 弹窗内容 */}
      <div className="flex min-h-full items-center justify-center p-4">
        <div className="relative bg-white rounded-2xl shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
          {/* 头部 */}
          <div className="sticky top-0 bg-white border-b px-6 py-4 flex items-center justify-between">
            <div>
              <h2 className="text-xl font-bold text-gray-900">升级会员</h2>
              <p className="text-sm text-gray-500 mt-1">选择适合您的会员计划</p>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* 计费周期切换 */}
          <div className="px-6 py-4 flex justify-center">
            <div className="bg-gray-100 rounded-lg p-1 inline-flex">
              <button
                onClick={() => setBillingCycle('monthly')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  billingCycle === 'monthly'
                    ? 'bg-white text-gray-900 shadow'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                月付
              </button>
              <button
                onClick={() => setBillingCycle('yearly')}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  billingCycle === 'yearly'
                    ? 'bg-white text-gray-900 shadow'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                年付 <span className="text-green-600 text-xs">省 17%</span>
              </button>
            </div>
          </div>

          {/* 等级卡片 */}
          <div className="px-6 pb-6">
            {loading ? (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="animate-pulse bg-gray-100 rounded-xl h-80"></div>
                ))}
              </div>
            ) : (
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {tiers.filter(t => t.tier !== 'free').map((tier) => {
                  const tierKey = tier.tier as MembershipTier;
                  const isCurrentTier = tierKey === currentTier;
                  const isUpgrade = tierOrder.indexOf(tierKey) > currentTierIndex;
                  const price = billingCycle === 'yearly' 
                    ? Math.round((tier.price || 0) * 10) 
                    : tier.price || 0;

                  return (
                    <div
                      key={tier.tier}
                      className={`relative rounded-xl border-2 p-6 transition-all ${
                        tierKey === 'pro'
                          ? 'border-purple-500 shadow-lg shadow-purple-100'
                          : 'border-gray-200 hover:border-gray-300'
                      } ${isCurrentTier ? 'bg-gray-50' : 'bg-white'}`}
                    >
                      {/* 推荐标签 */}
                      {tierKey === 'pro' && (
                        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                          <span className="bg-gradient-to-r from-purple-500 to-pink-500 text-white text-xs font-medium px-3 py-1 rounded-full">
                            最受欢迎
                          </span>
                        </div>
                      )}

                      {/* 等级信息 */}
                      <div className="text-center mb-4">
                        <span className="text-3xl">{TIER_ICONS[tierKey]}</span>
                        <h3 className="text-lg font-bold text-gray-900 mt-2">{tier.name}</h3>
                        <div className="mt-2">
                          <span className="text-3xl font-bold text-gray-900">¥{price}</span>
                          <span className="text-gray-500 text-sm">
                            /{billingCycle === 'yearly' ? '年' : '月'}
                          </span>
                        </div>
                      </div>

                      {/* 功能列表 */}
                      <ul className="space-y-2 mb-6">
                        {TIER_FEATURES[tierKey].map((feature, idx) => (
                          <li key={idx} className="flex items-start gap-2 text-sm">
                            <svg className="w-5 h-5 text-green-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                            </svg>
                            <span className="text-gray-600">{feature}</span>
                          </li>
                        ))}
                      </ul>

                      {/* 操作按钮 */}
                      <button
                        onClick={() => handleUpgrade(tierKey)}
                        disabled={isCurrentTier || !isUpgrade}
                        className={`w-full py-2 px-4 rounded-lg font-medium transition-colors ${
                          isCurrentTier
                            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                            : !isUpgrade
                            ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                            : tierKey === 'pro'
                            ? 'bg-gradient-to-r from-purple-500 to-pink-500 text-white hover:from-purple-600 hover:to-pink-600'
                            : 'bg-gray-900 text-white hover:bg-gray-800'
                        }`}
                      >
                        {isCurrentTier ? '当前方案' : !isUpgrade ? '已拥有' : '立即升级'}
                      </button>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* 底部说明 */}
          <div className="px-6 py-4 bg-gray-50 border-t text-center text-sm text-gray-500">
            <p>所有方案均支持 7 天无理由退款 · 随时可取消订阅</p>
          </div>
        </div>
      </div>
    </div>
  );
}
