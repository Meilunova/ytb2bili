'use client';

import { useState, useEffect } from 'react';
import { membershipApi } from '@/lib/api';
import type { BoostPackStatus, BoostPackType } from '@/types';

interface BoostPackCardProps {
  onPurchaseSuccess?: () => void;
}

const BOOST_PACKS: { type: BoostPackType; name: string; videos: number; price: number; validDays: number }[] = [
  { type: 'small', name: '小加油包', videos: 10, price: 9.9, validDays: 7 },
  { type: 'medium', name: '中加油包', videos: 30, price: 19.9, validDays: 15 },
  { type: 'large', name: '大加油包', videos: 100, price: 49.9, validDays: 30 },
];

export default function BoostPackCard({ onPurchaseSuccess }: BoostPackCardProps) {
  const [status, setStatus] = useState<BoostPackStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [purchasing, setPurchasing] = useState<BoostPackType | null>(null);
  const [showPurchase, setShowPurchase] = useState(false);

  useEffect(() => {
    fetchStatus();
  }, []);

  const fetchStatus = async () => {
    try {
      const res = await membershipApi.getBoostPackStatus();
      if (res.code === 0) {
        setStatus(res.data);
      }
    } catch (err) {
      console.error('获取加油包状态失败:', err);
    } finally {
      setLoading(false);
    }
  };

  const handlePurchase = async (packType: BoostPackType) => {
    try {
      setPurchasing(packType);
      const res = await membershipApi.purchaseBoostPack({ pack_type: packType });
      if (res.code === 0) {
        alert(`购买成功！获得 ${res.data.videos_added} 个视频配额`);
        fetchStatus();
        onPurchaseSuccess?.();
        setShowPurchase(false);
      } else {
        alert(res.message || '购买失败');
      }
    } catch (err: any) {
      alert(err.message || '购买失败');
    } finally {
      setPurchasing(null);
    }
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow p-4 animate-pulse">
        <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
        <div className="h-4 bg-gray-200 rounded w-2/3"></div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow overflow-hidden">
      {/* 头部 */}
      <div className="px-4 py-3 bg-gradient-to-r from-orange-400 to-red-400">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2 text-white">
            <span className="text-2xl">🚀</span>
            <h3 className="font-semibold">加油包</h3>
          </div>
          {!showPurchase && (
            <button
              onClick={() => setShowPurchase(true)}
              className="px-3 py-1 text-sm bg-white text-orange-500 rounded-full hover:bg-orange-50 transition-colors"
            >
              购买
            </button>
          )}
        </div>
      </div>

      {/* 当前状态 */}
      <div className="px-4 py-3">
        {status?.has_pack ? (
          <div className="space-y-2">
            <div className="flex justify-between items-center">
              <span className="text-gray-600">剩余配额</span>
              <span className="text-xl font-bold text-orange-600">
                {status.videos_remaining} 个视频
              </span>
            </div>
            <div className="flex justify-between items-center text-sm">
              <span className="text-gray-500">有效期</span>
              <span className="text-gray-600">
                {status.days_remaining > 0 ? `${status.days_remaining} 天` : '已过期'}
              </span>
            </div>
          </div>
        ) : (
          <div className="text-center py-2">
            <p className="text-gray-500 text-sm">暂无加油包</p>
            <p className="text-xs text-gray-400 mt-1">购买加油包可突破每日配额限制</p>
          </div>
        )}
      </div>

      {/* 购买选项 */}
      {showPurchase && (
        <div className="px-4 py-3 border-t bg-gray-50">
          <div className="flex justify-between items-center mb-3">
            <span className="text-sm font-medium text-gray-700">选择加油包</span>
            <button
              onClick={() => setShowPurchase(false)}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          <div className="space-y-2">
            {BOOST_PACKS.map((pack) => (
              <button
                key={pack.type}
                onClick={() => handlePurchase(pack.type)}
                disabled={purchasing !== null}
                className={`w-full flex items-center justify-between p-3 rounded-lg border transition-colors ${
                  purchasing === pack.type
                    ? 'bg-orange-50 border-orange-300'
                    : 'bg-white border-gray-200 hover:border-orange-300 hover:bg-orange-50'
                }`}
              >
                <div className="text-left">
                  <div className="font-medium text-gray-900">{pack.name}</div>
                  <div className="text-xs text-gray-500">
                    {pack.videos} 个视频 · {pack.validDays} 天有效
                  </div>
                </div>
                <div className="text-right">
                  {purchasing === pack.type ? (
                    <span className="text-orange-500">购买中...</span>
                  ) : (
                    <span className="text-lg font-bold text-orange-600">¥{pack.price}</span>
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
