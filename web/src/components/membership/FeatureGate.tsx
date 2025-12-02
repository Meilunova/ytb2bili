'use client';

import { useState, useEffect, ReactNode } from 'react';
import { membershipApi } from '@/lib/api';
import type { FeatureCheckResult, MembershipTier } from '@/types';
import UpgradeModal from './UpgradeModal';

interface FeatureGateProps {
  feature: string;
  children: ReactNode;
  fallback?: ReactNode;
  showUpgradePrompt?: boolean;
}

const FEATURE_NAMES: Record<string, string> = {
  ai_translation: 'AI 字幕翻译',
  translation_optimize: '翻译质量优化',
  ai_title_generation: 'AI 标题生成',
  gemini_video_analysis: 'Gemini 视频分析',
  auto_upload: '自动上传',
  priority_queue: '优先队列',
  api_access: 'API 访问',
  custom_template: '自定义模板',
  data_export: '数据导出',
  team_collaboration: '团队协作',
};

export default function FeatureGate({ 
  feature, 
  children, 
  fallback,
  showUpgradePrompt = true 
}: FeatureGateProps) {
  const [checkResult, setCheckResult] = useState<FeatureCheckResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [showUpgrade, setShowUpgrade] = useState(false);

  useEffect(() => {
    checkFeature();
  }, [feature]);

  const checkFeature = async () => {
    try {
      setLoading(true);
      const res = await membershipApi.checkFeature(feature);
      if (res.code === 0) {
        setCheckResult(res.data);
      }
    } catch (err) {
      console.error('检查功能权限失败:', err);
      // 默认允许（避免阻塞用户）
      setCheckResult({ feature, allowed: true });
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="h-8 bg-gray-200 rounded w-full"></div>
      </div>
    );
  }

  // 功能可用，显示子组件
  if (checkResult?.allowed) {
    return <>{children}</>;
  }

  // 功能不可用
  const featureName = FEATURE_NAMES[feature] || feature;

  // 如果有自定义 fallback，使用它
  if (fallback) {
    return <>{fallback}</>;
  }

  // 默认的升级提示
  if (showUpgradePrompt) {
    return (
      <>
        <div className="relative">
          {/* 模糊的子组件 */}
          <div className="opacity-50 pointer-events-none blur-sm">
            {children}
          </div>
          
          {/* 升级提示覆盖层 */}
          <div className="absolute inset-0 flex items-center justify-center bg-white/80 rounded-lg">
            <div className="text-center p-4">
              <div className="text-4xl mb-2">🔒</div>
              <h4 className="font-medium text-gray-900 mb-1">{featureName}</h4>
              <p className="text-sm text-gray-500 mb-3">{checkResult?.reason}</p>
              <button
                onClick={() => setShowUpgrade(true)}
                className="px-4 py-2 bg-gradient-to-r from-purple-500 to-pink-500 text-white text-sm rounded-lg hover:from-purple-600 hover:to-pink-600 transition-all"
              >
                升级到 {checkResult?.suggestion === 'basic' ? '基础版' : 
                        checkResult?.suggestion === 'pro' ? '专业版' : 
                        checkResult?.suggestion === 'enterprise' ? '企业版' : '更高版本'}
              </button>
            </div>
          </div>
        </div>

        <UpgradeModal 
          isOpen={showUpgrade} 
          onClose={() => setShowUpgrade(false)} 
        />
      </>
    );
  }

  // 不显示任何内容
  return null;
}
