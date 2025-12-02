'use client';

import { useState, useEffect } from 'react';
import { membershipApi } from '@/lib/api';
import type { QuotaInfo } from '@/types';

interface QuotaDisplayProps {
  compact?: boolean;
  showBoostPack?: boolean;
  onQuotaExhausted?: () => void;
}

export default function QuotaDisplay({ 
  compact = false, 
  showBoostPack = true,
  onQuotaExhausted 
}: QuotaDisplayProps) {
  const [quota, setQuota] = useState<QuotaInfo | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchQuota();
  }, []);

  const fetchQuota = async () => {
    try {
      const res = await membershipApi.getQuotaInfo();
      if (res.code === 0) {
        setQuota(res.data);
        if (res.data.total_remaining === 0 && !res.data.is_unlimited) {
          onQuotaExhausted?.();
        }
      }
    } catch (err) {
      console.error('获取配额失败:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="animate-pulse">
        <div className="h-4 bg-gray-200 rounded w-20"></div>
      </div>
    );
  }

  if (!quota) {
    return null;
  }

  // 紧凑模式 - 用于顶部导航栏
  if (compact) {
    if (quota.is_unlimited) {
      return (
        <div className="flex items-center gap-1 text-sm text-green-600">
          <span>∞</span>
          <span>无限</span>
        </div>
      );
    }

    const isLow = quota.total_remaining <= 2;
    const isEmpty = quota.total_remaining === 0;

    return (
      <div className={`flex items-center gap-1 text-sm ${
        isEmpty ? 'text-red-600' : isLow ? 'text-orange-600' : 'text-gray-600'
      }`}>
        <span className="font-medium">{quota.total_remaining}</span>
        <span className="text-gray-400">/</span>
        <span>{quota.daily_limit}</span>
        {quota.boost_pack_remaining > 0 && (
          <span className="text-orange-500 ml-1">+{quota.boost_pack_remaining}</span>
        )}
      </div>
    );
  }

  // 完整模式
  return (
    <div className="bg-white rounded-lg shadow p-4">
      <h4 className="text-sm font-medium text-gray-700 mb-3">配额使用情况</h4>
      
      {quota.is_unlimited ? (
        <div className="text-center py-4">
          <span className="text-4xl">∞</span>
          <p className="text-green-600 mt-2">无限制使用</p>
        </div>
      ) : (
        <>
          {/* 每日配额 */}
          <div className="mb-4">
            <div className="flex justify-between text-sm mb-1">
              <span className="text-gray-600">每日配额</span>
              <span>
                <span className="font-medium text-blue-600">{quota.daily_remaining}</span>
                <span className="text-gray-400"> / {quota.daily_limit}</span>
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`h-2 rounded-full transition-all ${
                  quota.daily_remaining === 0 ? 'bg-red-500' : 
                  quota.daily_remaining <= 2 ? 'bg-orange-500' : 'bg-blue-500'
                }`}
                style={{
                  width: `${Math.max(0, (quota.daily_remaining / quota.daily_limit) * 100)}%`,
                }}
              ></div>
            </div>
          </div>

          {/* 加油包 */}
          {showBoostPack && (
            <div className="pt-3 border-t">
              <div className="flex justify-between items-center">
                <div className="flex items-center gap-2">
                  <span className="text-lg">🚀</span>
                  <span className="text-sm text-gray-600">加油包</span>
                </div>
                <span className={`font-medium ${
                  quota.boost_pack_remaining > 0 ? 'text-orange-600' : 'text-gray-400'
                }`}>
                  {quota.boost_pack_remaining > 0 
                    ? `${quota.boost_pack_remaining} 个视频` 
                    : '未购买'}
                </span>
              </div>
            </div>
          )}

          {/* 总剩余 */}
          <div className="mt-4 pt-3 border-t">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-700">总可用</span>
              <span className={`text-lg font-bold ${
                quota.total_remaining === 0 ? 'text-red-600' : 'text-green-600'
              }`}>
                {quota.total_remaining}
              </span>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
