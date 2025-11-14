import { useState, useEffect, useCallback } from 'react';

interface UserInfo {
  id: string;
  name: string;
  mid: string;
  avatar?: string;
}

export function useAuth() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);

  // 获取API基础URL
  const getApiBaseUrl = () => {
    if (typeof window !== 'undefined') {
      const { protocol, hostname, port } = window.location;
      // 在embed模式下，前端和后端运行在同一个端口
      return `${protocol}//${hostname}${port ? ':' + port : ''}`;
    }
    return 'http://localhost:8096';
  };

  const checkAuthStatus = useCallback(async () => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/auth/status`);
      const data = await response.json();
      
      console.log('Auth status response:', data); // 调试日志
      
      if (data.code === 0 && data.is_logged_in && data.user) {
        console.log('User is logged in:', data.user); // 调试日志
        setUser(data.user);
      } else {
        console.log('User is not logged in'); // 调试日志
      }
    } catch (error) {
      console.error('检查登录状态失败:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 检查登录状态
  useEffect(() => {
    checkAuthStatus();
  }, [checkAuthStatus]);

  const handleLoginSuccess = (userData: UserInfo) => {
    setUser(userData);
  };

  const handleRefreshStatus = async () => {
    // 重新检查登录状态，用于二维码登录成功后的状态同步
    await checkAuthStatus();
  };

  const handleLogout = async () => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      await fetch(`${apiBaseUrl}/api/v1/auth/logout`, { method: 'POST' });
      setUser(null);
    } catch (error) {
      console.error('登出失败:', error);
    }
  };

  return {
    user,
    loading,
    handleLoginSuccess,
    handleRefreshStatus,
    handleLogout,
  };
}