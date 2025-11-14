"use client";

import Link from 'next/link';
import AppLayout from '@/components/layout/AppLayout';
import QRLogin from '@/components/auth/QRLogin';
import { useAuth } from '@/hooks/useAuth';
import { BarChart3, Clock, Download, ExternalLink, Chrome, Settings } from 'lucide-react';
import { useState } from 'react';

export default function HomePage() {
  const { user, loading, handleLoginSuccess, handleRefreshStatus, handleLogout } = useAuth();
  const [isDownloading, setIsDownloading] = useState(false);

  const handleDownloadExtension = async () => {
    setIsDownloading(true);
    try {
      // 获取最新版本信息
      const response = await fetch('https://api.github.com/repos/difyz9/bili-up-extension/releases/latest');
      const release = await response.json();
      
      if (release.assets && release.assets.length > 0) {
        // 查找.zip文件
        const zipAsset = release.assets.find((asset: any) => asset.name.endsWith('.zip'));
        if (zipAsset) {
          // 创建下载链接
          const link = document.createElement('a');
          link.href = zipAsset.browser_download_url;
          link.download = zipAsset.name;
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
        } else {
          // 如果没有找到zip文件，打开releases页面
          window.open('https://github.com/difyz9/bili-up-extension/releases/latest', '_blank');
        }
      } else {
        window.open('https://github.com/difyz9/bili-up-extension/releases/latest', '_blank');
      }
    } catch (error) {
      console.error('下载失败:', error);
      // 如果API调用失败，打开releases页面
      window.open('https://github.com/difyz9/bili-up-extension/releases/latest', '_blank');
    } finally {
      setIsDownloading(false);
    }
  };

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
        {/* 插件下载导引 */}
        <div className="bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg border border-purple-200 p-6">
          <div className="flex items-start space-x-4">
            <div className="flex-shrink-0">
              <Chrome className="w-8 h-8 text-purple-600" />
            </div>
            <div className="flex-1">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                安装浏览器插件获取视频信息
              </h3>
              <p className="text-gray-600 mb-4">
                为了获取 B 站视频信息和更好的使用体验，建议安装我们的浏览器插件。插件可以帮助您：
              </p>
              <ul className="text-sm text-gray-600 mb-4 space-y-1">
                <li>• 自动获取视频标题、描述和封面</li>
                <li>• 快速导入视频链接到上传队列</li>
                <li>• 同步浏览器中的视频收藏</li>
              </ul>
              <div className="flex flex-col sm:flex-row gap-3">
                <button
                  onClick={handleDownloadExtension}
                  disabled={isDownloading}
                  className="flex items-center justify-center px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isDownloading ? (
                    <>
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"></div>
                      下载中...
                    </>
                  ) : (
                    <>
                      <Download className="w-4 h-4 mr-2" />
                      下载最新插件
                    </>
                  )}
                </button>
                <a
                  href="https://github.com/difyz9/bili-up-extension"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center justify-center px-4 py-2 border border-purple-300 text-purple-700 rounded-md hover:bg-purple-50 transition-colors"
                >
                  <ExternalLink className="w-4 h-4 mr-2" />
                  查看源码
                </a>
              </div>
              <div className="mt-3 text-xs text-gray-500">
                下载后请解压文件，然后在浏览器扩展管理页面选择&ldquo;加载已解压的扩展程序&rdquo;
              </div>
            </div>
          </div>
        </div>

        {/* 主要功能区域 */}
        <div className="bg-white rounded-lg shadow-md p-8">
          <div className="text-center">
            <h2 className="text-2xl font-semibold text-gray-900 mb-4">
              欢迎使用 Bili-Up Web
            </h2>
            <p className="text-gray-600 mb-6">
              请选择左侧导航菜单来访问不同功能模块
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-lg mx-auto">
              <Link
                href="/dashboard"
                className="flex items-center justify-center px-4 py-3 bg-blue-50 text-blue-700 rounded-lg hover:bg-blue-100 transition-colors"
              >
                <BarChart3 className="w-5 h-5 mr-2" />
                任务队列
              </Link>
              <Link
                href="/schedule"
                className="flex items-center justify-center px-4 py-3 bg-green-50 text-green-700 rounded-lg hover:bg-green-100 transition-colors"
              >
                <Clock className="w-5 h-5 mr-2" />
                定时上传
              </Link>
              <Link
                href="/extension"
                className="flex items-center justify-center px-4 py-3 bg-purple-50 text-purple-700 rounded-lg hover:bg-purple-100 transition-colors"
              >
                <Chrome className="w-5 h-5 mr-2" />
                浏览器插件
              </Link>
              <Link
                href="/settings"
                className="flex items-center justify-center px-4 py-3 bg-gray-50 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <Settings className="w-5 h-5 mr-2" />
                设置
              </Link>
            </div>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}
