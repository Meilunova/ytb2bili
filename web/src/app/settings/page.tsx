"use client";

import { useState, useEffect } from 'react';
import AppLayout from '@/components/layout/AppLayout';
import QRLogin from '@/components/auth/QRLogin';
import AIModelSettings from '@/components/settings/AIModelSettings';
import { useAuth } from '@/hooks/useAuth';
import { Settings, Bot, Upload } from 'lucide-react';

export default function SettingsPage() {
  const { user, loading, handleLoginSuccess, handleRefreshStatus, handleLogout } = useAuth();
  const [activeTab, setActiveTab] = useState<'general' | 'ai'>('general');
  const [autoUpload, setAutoUpload] = useState<boolean>(() => {
    if (typeof window !== 'undefined') {
      try {
        const v = localStorage.getItem('biliup:autoUpload');
        return v === '1';
      } catch {
        return false;
      }
    }
    return false;
  });

  useEffect(() => {
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem('biliup:autoUpload', autoUpload ? '1' : '0');
      } catch {
        // ignore
      }
    }
  }, [autoUpload]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="inline-block w-8 h-8 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4"></div>
          <p className="text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="container mx-auto px-4 py-16">
          <div className="max-w-md mx-auto">
            <div className="text-center mb-8">
              <h1 className="text-3xl font-bold text-gray-900 mb-2">
                Bili-Up Web
              </h1>
              <p className="text-gray-600">
                Bilibili 视频管理平台
              </p>
            </div>
            
            <div className="bg-white rounded-lg shadow-lg">
              <QRLogin 
                onLoginSuccess={handleLoginSuccess}
                onRefreshStatus={handleRefreshStatus}
              />
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <AppLayout user={user} onLogout={handleLogout}>
      <div className="space-y-6">
        {/* 标签页导航 */}
        <div className="bg-white rounded-lg shadow-md">
          <div className="border-b border-gray-200">
            <nav className="flex -mb-px">
              <button
                onClick={() => setActiveTab('general')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'general'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <Settings className="w-4 h-4 inline mr-2" />
                通用设置
              </button>
              <button
                onClick={() => setActiveTab('ai')}
                className={`px-6 py-4 text-sm font-medium border-b-2 transition-colors ${
                  activeTab === 'ai'
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <Bot className="w-4 h-4 inline mr-2" />
                AI 大模型
              </button>
            </nav>
          </div>
        </div>

        {/* 通用设置 */}
        {activeTab === 'general' && (
          <div className="bg-white rounded-lg shadow-md">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center space-x-3">
                <Settings className="w-5 h-5 text-gray-600" />
                <h2 className="text-lg font-medium text-gray-900">通用设置</h2>
              </div>
            </div>

            <div className="p-6">
              <div className="space-y-4">
                <label className="flex items-center justify-between bg-gray-50 p-4 rounded-md">
                  <div>
                    <div className="text-sm font-medium flex items-center">
                      <Upload className="w-4 h-4 mr-2 text-gray-500" />
                      自动上传
                    </div>
                    <div className="text-xs text-gray-500 mt-1">视频提交后自动开始上传任务</div>
                  </div>
                  <div className="relative">
                    <input
                      type="checkbox"
                      checked={autoUpload}
                      onChange={(e) => setAutoUpload(e.target.checked)}
                      className="sr-only"
                    />
                    <div 
                      onClick={() => setAutoUpload(!autoUpload)}
                      className={`w-10 h-6 rounded-full cursor-pointer transition-colors ${autoUpload ? 'bg-blue-600' : 'bg-gray-300'}`}
                    >
                      <div className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform ${autoUpload ? 'translate-x-4' : ''}`} />
                    </div>
                  </div>
                </label>

                <div className="bg-blue-50 p-4 rounded-md">
                  <div className="text-sm text-blue-800">
                    <strong>提示：</strong> 更多通用设置项将在后续版本中添加。
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* AI 大模型设置 */}
        {activeTab === 'ai' && <AIModelSettings />}
      </div>
    </AppLayout>
  );
}
