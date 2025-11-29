"use client";

import { useState, useEffect, useCallback } from 'react';
import { 
  Bot, 
  Settings2, 
  TestTube, 
  Check, 
  X, 
  Loader2, 
  ChevronDown,
  Zap,
  Globe,
  Key,
  Clock,
  Thermometer,
  Trash2,
  RefreshCw
} from 'lucide-react';

// 提供商信息接口
interface ProviderInfo {
  id: string;
  name: string;
  description: string;
  base_url: string;
  default_model: string;
}

// 配置接口
interface OpenAICompatibleConfig {
  enabled: boolean;
  provider: string;
  api_key: string;
  base_url: string;
  model: string;
  timeout: number;
  max_tokens: number;
  temperature: number;
}

// 测试结果接口
interface TestResult {
  success: boolean;
  message: string;
  response?: string;
  latency_ms?: number;
}

// AI服务状态接口
interface AIServiceStatus {
  provider: string;
  name: string;
  enabled: boolean;
  available: boolean;
  model?: string;
  base_url?: string;
  is_primary: boolean;
  last_error?: string;
}

interface AIServicesStatusResponse {
  services: AIServiceStatus[];
  primary_provider: string;
  has_available: boolean;
}

// Gemini原生配置接口（用于元数据生成）
interface GeminiConfig {
  enabled: boolean;
  api_key: string;
  api_keys: string[];
  api_keys_count: number;
  model: string;
  timeout: number;
  max_tokens: number;
  use_for_metadata: boolean;
  analyze_video: boolean;
  video_sample_frames: number;
}

// 获取API基础URL
const getApiBaseUrl = () => {
  if (typeof window !== 'undefined') {
    const { protocol, hostname, port } = window.location;
    return `${protocol}//${hostname}${port ? ':' + port : ''}`;
  }
  return 'http://localhost:8096';
};

export default function AIModelSettings() {
  // 状态
  const [config, setConfig] = useState<OpenAICompatibleConfig>({
    enabled: false,
    provider: 'openai',
    api_key: '',
    base_url: 'https://api.openai.com/v1',
    model: 'gpt-3.5-turbo',
    timeout: 60,
    max_tokens: 4000,
    temperature: 0.7,
  });
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [testResult, setTestResult] = useState<TestResult | null>(null);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [apiKeyInput, setApiKeyInput] = useState('');
  const [hasChanges, setHasChanges] = useState(false);
  const [servicesStatus, setServicesStatus] = useState<AIServicesStatusResponse | null>(null);

  // Gemini原生配置状态（用于元数据生成）
  const [geminiConfig, setGeminiConfig] = useState<GeminiConfig>({
    enabled: false,
    api_key: '',
    api_keys: [],
    api_keys_count: 0,
    model: 'gemini-2.0-flash',
    timeout: 120,
    max_tokens: 8000,
    use_for_metadata: true,
    analyze_video: true,
    video_sample_frames: 0,
  });
  const [geminiApiKeysInput, setGeminiApiKeysInput] = useState(''); // 多个 API Key，用换行分隔
  const [geminiHasChanges, setGeminiHasChanges] = useState(false);
  const [geminiSaving, setGeminiSaving] = useState(false);
  const [geminiClearing, setGeminiClearing] = useState(false);
  const [geminiRefreshing, setGeminiRefreshing] = useState(false);
  const [geminiValidating, setGeminiValidating] = useState(false);
  const [geminiValidationResults, setGeminiValidationResults] = useState<{
    total_keys: number;
    valid_keys: number;
    invalid_keys: number;
    results: Array<{key: string; index: number; valid: boolean; message: string}>;
  } | null>(null);

  // 加载配置
  const loadConfig = useCallback(async () => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/openai-compatible`);
      const data = await response.json();
      if (data.code === 200 && data.data) {
        setConfig(data.data);
      }
    } catch (error) {
      console.error('加载配置失败:', error);
    }
  }, []);

  // 加载提供商列表
  const loadProviders = useCallback(async () => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/openai-compatible/providers`);
      const data = await response.json();
      if (data.code === 200 && data.data) {
        setProviders(data.data);
      }
    } catch (error) {
      console.error('加载提供商列表失败:', error);
    }
  }, []);

  // 加载AI服务状态
  const loadServicesStatus = useCallback(async () => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/ai-services/status`);
      const data = await response.json();
      if (data.code === 200 && data.data) {
        setServicesStatus(data.data);
      }
    } catch (error) {
      console.error('加载AI服务状态失败:', error);
    }
  }, []);

  // 加载Gemini配置
  const loadGeminiConfig = useCallback(async () => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/gemini`);
      const data = await response.json();
      if (data.code === 200 && data.data) {
        setGeminiConfig(data.data);
      }
    } catch (error) {
      console.error('加载Gemini配置失败:', error);
    }
  }, []);

  // 初始化加载
  useEffect(() => {
    Promise.all([loadConfig(), loadProviders(), loadServicesStatus(), loadGeminiConfig()]).finally(() => setLoading(false));
  }, [loadConfig, loadProviders, loadServicesStatus, loadGeminiConfig]);

  // 保存配置
  const saveConfig = async () => {
    setSaving(true);
    try {
      const apiBaseUrl = getApiBaseUrl();
      const updateData: Partial<OpenAICompatibleConfig> & { api_key?: string } = {
        enabled: config.enabled,
        provider: config.provider,
        base_url: config.base_url,
        model: config.model,
        timeout: config.timeout,
        max_tokens: config.max_tokens,
        temperature: config.temperature,
      };
      
      // 只有当用户输入了新的API Key时才更新
      if (apiKeyInput) {
        updateData.api_key = apiKeyInput;
      }

      const response = await fetch(`${apiBaseUrl}/api/v1/config/openai-compatible`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updateData),
      });
      
      const data = await response.json();
      if (data.code === 200) {
        setConfig(data.data);
        setApiKeyInput('');
        setHasChanges(false);
        // 刷新服务状态
        await loadServicesStatus();
        alert('配置保存成功！');
      } else {
        alert('保存失败: ' + data.message);
      }
    } catch (error) {
      console.error('保存配置失败:', error);
      alert('保存失败，请检查网络连接');
    } finally {
      setSaving(false);
    }
  };

  // 测试API连接
  const testConnection = async () => {
    setTesting(true);
    setTestResult(null);
    try {
      const apiBaseUrl = getApiBaseUrl();
      const testData = {
        provider: config.provider,
        api_key: apiKeyInput || config.api_key,
        base_url: config.base_url,
        model: config.model,
        timeout: config.timeout,
        temperature: config.temperature,
      };

      // 如果没有API Key，提示用户
      if (!testData.api_key || testData.api_key.includes('...')) {
        setTestResult({
          success: false,
          message: '请先输入API Key',
        });
        setTesting(false);
        return;
      }

      const response = await fetch(`${apiBaseUrl}/api/v1/config/openai-compatible/test`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(testData),
      });
      
      const data = await response.json();
      if (data.code === 200) {
        setTestResult(data.data);
      } else {
        setTestResult({
          success: false,
          message: data.message || '测试失败',
        });
      }
    } catch (error) {
      console.error('测试连接失败:', error);
      setTestResult({
        success: false,
        message: '网络请求失败，请检查网络连接',
      });
    } finally {
      setTesting(false);
    }
  };

  // 切换提供商
  const handleProviderChange = (providerId: string) => {
    const provider = providers.find(p => p.id === providerId);
    if (provider) {
      setConfig(prev => ({
        ...prev,
        provider: providerId,
        base_url: provider.base_url || prev.base_url,
        model: provider.default_model || prev.model,
      }));
      setHasChanges(true);
    }
  };

  // 更新配置字段
  const updateConfig = (field: keyof OpenAICompatibleConfig, value: string | number | boolean) => {
    setConfig(prev => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  // 设置首选AI服务
  const setPrimaryService = async (provider: string) => {
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/ai-services/primary`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ provider }),
      });
      
      const data = await response.json();
      if (data.code === 200) {
        // 刷新服务状态
        await loadServicesStatus();
        alert(`已将 "${servicesStatus?.services.find(s => s.provider === provider)?.name || provider}" 设为首选服务`);
      } else {
        alert('设置失败: ' + data.message);
      }
    } catch (error) {
      console.error('设置首选服务失败:', error);
      alert('设置失败，请检查网络连接');
    }
  };

  // 更新Gemini配置字段
  const updateGeminiConfig = (field: keyof GeminiConfig, value: string | number | boolean) => {
    setGeminiConfig(prev => ({ ...prev, [field]: value }));
    setGeminiHasChanges(true);
  };

  // 保存Gemini配置
  const saveGeminiConfig = async () => {
    setGeminiSaving(true);
    try {
      const apiBaseUrl = getApiBaseUrl();
      const updateData: Partial<GeminiConfig> & { api_keys?: string[] } = {
        enabled: geminiConfig.enabled,
        model: geminiConfig.model,
        timeout: geminiConfig.timeout,
        max_tokens: geminiConfig.max_tokens,
        use_for_metadata: geminiConfig.use_for_metadata,
        analyze_video: geminiConfig.analyze_video,
        video_sample_frames: geminiConfig.video_sample_frames,
      };
      
      // 解析多个 API Key（用换行分隔）
      if (geminiApiKeysInput.trim()) {
        const keys = geminiApiKeysInput
          .split('\n')
          .map(k => k.trim())
          .filter(k => k.length > 0);
        if (keys.length > 0) {
          updateData.api_keys = keys;
        }
      }

      const response = await fetch(`${apiBaseUrl}/api/v1/config/gemini`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updateData),
      });
      
      const data = await response.json();
      if (data.code === 200) {
        setGeminiConfig(data.data);
        setGeminiApiKeysInput('');
        setGeminiHasChanges(false);
        // 刷新服务状态
        await loadServicesStatus();
        alert(`Gemini 配置保存成功！(${data.data.api_keys_count} 个 API Key)`);
      } else {
        alert('保存失败: ' + data.message);
      }
    } catch (error) {
      console.error('保存Gemini配置失败:', error);
      alert('保存失败，请检查网络连接');
    } finally {
      setGeminiSaving(false);
    }
  };

  // 清空Gemini API Keys
  const clearGeminiApiKeys = async () => {
    if (!confirm('确定要清空所有 Gemini API Keys 吗？此操作不可恢复。')) {
      return;
    }
    
    setGeminiClearing(true);
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/gemini`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          api_keys: [], // 发送空数组来清空
          clear_api_keys: true, // 明确标记要清空
        }),
      });
      
      const data = await response.json();
      if (data.code === 200) {
        setGeminiConfig(data.data);
        setGeminiApiKeysInput('');
        // 刷新服务状态
        await loadServicesStatus();
        alert('已清空所有 Gemini API Keys');
      } else {
        alert('清空失败: ' + data.message);
      }
    } catch (error) {
      console.error('清空Gemini API Keys失败:', error);
      alert('清空失败，请检查网络连接');
    } finally {
      setGeminiClearing(false);
    }
  };

  // 刷新Gemini可用模型列表
  const refreshGeminiModels = async () => {
    setGeminiRefreshing(true);
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/gemini/models`);
      const data = await response.json();
      if (data.code === 200 && data.data?.models) {
        alert(`可用模型: ${data.data.models.join(', ')}`);
      } else {
        alert('获取模型列表失败: ' + (data.message || '未知错误'));
      }
    } catch (error) {
      console.error('获取Gemini模型列表失败:', error);
      alert('获取失败，请检查网络连接或 API Key 是否有效');
    } finally {
      setGeminiRefreshing(false);
    }
  };

  // 验证Gemini API Keys
  const validateGeminiApiKeys = async () => {
    setGeminiValidating(true);
    setGeminiValidationResults(null);
    try {
      const apiBaseUrl = getApiBaseUrl();
      const response = await fetch(`${apiBaseUrl}/api/v1/config/gemini/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
      const data = await response.json();
      if (data.code === 200 && data.data) {
        setGeminiValidationResults(data.data);
        if (data.data.invalid_keys > 0) {
          alert(`验证完成！\n✅ 有效: ${data.data.valid_keys} 个\n❌ 无效: ${data.data.invalid_keys} 个\n\n建议清除无效的 API Key`);
        } else {
          alert(`验证完成！所有 ${data.data.valid_keys} 个 API Key 均有效 ✅`);
        }
      } else {
        alert('验证失败: ' + (data.message || '未知错误'));
      }
    } catch (error) {
      console.error('验证Gemini API Keys失败:', error);
      alert('验证失败，请检查网络连接');
    } finally {
      setGeminiValidating(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="w-6 h-6 animate-spin text-blue-500" />
        <span className="ml-2 text-gray-600">加载中...</span>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-md">
      {/* 标题栏 */}
      <div className="p-6 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <Bot className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-medium text-gray-900">AI 大模型配置</h2>
          </div>
          <label className="flex items-center space-x-2 cursor-pointer">
            <span className="text-sm text-gray-600">启用</span>
            <div className="relative">
              <input
                type="checkbox"
                checked={config.enabled}
                onChange={(e) => updateConfig('enabled', e.target.checked)}
                className="sr-only"
              />
              <div className={`w-10 h-6 rounded-full transition-colors ${config.enabled ? 'bg-blue-600' : 'bg-gray-300'}`}>
                <div className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform ${config.enabled ? 'translate-x-4' : ''}`} />
              </div>
            </div>
          </label>
        </div>
        <p className="mt-2 text-sm text-gray-500">
          配置 OpenAI 兼容的大语言模型 API，支持 OpenAI、DeepSeek、通义千问等多种服务
        </p>
      </div>

      <div className="p-6 space-y-6">
        {/* AI服务状态概览 - 点击选择首选服务 */}
        {servicesStatus && (
          <div className="bg-gray-50 rounded-lg p-4">
            <h3 className="text-sm font-medium text-gray-700 mb-3 flex items-center">
              <Zap className="w-4 h-4 mr-1" />
              AI服务状态
              <span className="ml-2 text-xs text-gray-500">（点击选择首选服务）</span>
            </h3>
            <div className="space-y-2">
              {servicesStatus.services.map((service) => (
                <div 
                  key={service.provider}
                  onClick={() => service.enabled && setPrimaryService(service.provider)}
                  className={`flex items-center justify-between p-3 rounded cursor-pointer transition-all ${
                    service.is_primary 
                      ? 'bg-blue-100 border-2 border-blue-400 shadow-sm' 
                      : service.enabled 
                        ? 'bg-white border border-gray-200 hover:border-blue-300 hover:bg-blue-50' 
                        : 'bg-gray-100 border border-gray-200 cursor-not-allowed opacity-60'
                  }`}
                >
                  <div className="flex items-center space-x-3">
                    <div className={`w-3 h-3 rounded-full ${
                      service.enabled && service.available ? 'bg-green-500' : 'bg-gray-300'
                    }`} />
                    <span className={`text-sm font-medium ${service.is_primary ? 'text-blue-700' : 'text-gray-700'}`}>
                      {service.name}
                    </span>
                    {service.is_primary && (
                      <span className="text-xs bg-blue-500 text-white px-2 py-0.5 rounded font-medium">首选</span>
                    )}
                  </div>
                  <div className="flex items-center space-x-2">
                    {service.model && (
                      <span className="text-xs text-gray-500">{service.model}</span>
                    )}
                    <span className={`text-xs px-2 py-0.5 rounded ${
                      service.enabled ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
                    }`}>
                      {service.enabled ? '已启用' : '未启用'}
                    </span>
                  </div>
                </div>
              ))}
            </div>
            {!servicesStatus.has_available && (
              <p className="mt-3 text-sm text-amber-600">
                ⚠️ 没有可用的AI服务，请配置至少一个AI服务
              </p>
            )}
            {servicesStatus.has_available && (
              <div className="mt-3 space-y-1">
                <p className="text-sm text-green-600">
                  ✓ 翻译服务: {servicesStatus.services.find(s => s.is_primary)?.name || servicesStatus.primary_provider}
                  <span className="text-gray-500 ml-2">（点击上方切换）</span>
                </p>
                {servicesStatus.services.find(s => s.provider === 'gemini')?.enabled ? (
                  <p className="text-sm text-blue-600">
                    ✓ 元数据生成: Gemini（原生多模态）
                    <span className="text-gray-500 ml-2">（固定使用，支持视频分析）</span>
                  </p>
                ) : (
                  <p className="text-sm text-amber-600">
                    ⚠️ 元数据生成: 需要配置 Gemini
                    <span className="text-gray-500 ml-2">（Gemini 具有多模态视频分析能力）</span>
                  </p>
                )}
              </div>
            )}
          </div>
        )}

        {/* Gemini 原生配置（用于元数据生成） */}
        <div className="bg-purple-50 rounded-lg p-4 border border-purple-200">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-sm font-medium text-purple-800 flex items-center">
              🔮 Gemini 原生配置（元数据生成专用）
            </h3>
            <label className="flex items-center space-x-2 cursor-pointer">
              <span className="text-sm text-purple-600">启用</span>
              <div className="relative">
                <input
                  type="checkbox"
                  checked={geminiConfig.enabled}
                  onChange={(e) => updateGeminiConfig('enabled', e.target.checked)}
                  className="sr-only"
                />
                <div className={`w-10 h-6 rounded-full transition-colors ${geminiConfig.enabled ? 'bg-purple-600' : 'bg-gray-300'}`}>
                  <div className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform ${geminiConfig.enabled ? 'translate-x-4' : ''}`} />
                </div>
              </div>
            </label>
          </div>
          
          <p className="text-xs text-purple-600 mb-4">
            Gemini 具有多模态视频分析能力，是生成高质量元数据的最佳选择。此配置独立于翻译服务。
          </p>

          <div className="space-y-4">
            {/* API Keys（多个，用于轮询） */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                <Key className="w-4 h-4 inline mr-1" />
                API Keys（支持多个，用于轮询）
              </label>
              <textarea
                value={geminiApiKeysInput}
                onChange={(e) => {
                  setGeminiApiKeysInput(e.target.value);
                  setGeminiHasChanges(true);
                }}
                placeholder={geminiConfig.api_keys_count > 0 
                  ? `当前已配置 ${geminiConfig.api_keys_count} 个 API Key\n输入新的 API Key 将替换现有配置\n每行一个 API Key` 
                  : '请输入 Gemini API Key\n每行一个，支持多个 Key 轮询使用'}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-purple-500 focus:border-purple-500 font-mono text-sm"
              />
              <div className="flex justify-between items-center mt-1">
                <p className="text-xs text-gray-500">
                  从 <a href="https://aistudio.google.com/app/apikey" target="_blank" rel="noopener noreferrer" className="text-purple-600 hover:underline">Google AI Studio</a> 获取，每行一个 Key
                </p>
                <div className="flex items-center space-x-2">
                  {geminiConfig.api_keys_count > 0 && (
                    <>
                      <span className="text-xs text-purple-600 font-medium">
                        已配置 {geminiConfig.api_keys_count} 个 Key
                      </span>
                      <button
                        onClick={validateGeminiApiKeys}
                        disabled={geminiValidating}
                        className="flex items-center px-2 py-1 text-xs text-blue-600 hover:text-blue-700 hover:bg-blue-50 rounded transition-colors disabled:opacity-50"
                        title="验证所有 API Keys 的有效性"
                      >
                        {geminiValidating ? (
                          <Loader2 className="w-3 h-3 animate-spin" />
                        ) : (
                          <TestTube className="w-3 h-3" />
                        )}
                        <span className="ml-1">验证</span>
                      </button>
                      <button
                        onClick={clearGeminiApiKeys}
                        disabled={geminiClearing}
                        className="flex items-center px-2 py-1 text-xs text-red-600 hover:text-red-700 hover:bg-red-50 rounded transition-colors disabled:opacity-50"
                        title="清空所有 API Keys"
                      >
                        {geminiClearing ? (
                          <Loader2 className="w-3 h-3 animate-spin" />
                        ) : (
                          <Trash2 className="w-3 h-3" />
                        )}
                        <span className="ml-1">清空</span>
                      </button>
                    </>
                  )}
                </div>
              </div>
            </div>

            {/* API Key 验证结果 */}
            {geminiValidationResults && (
              <div className={`rounded-md p-3 ${
                geminiValidationResults.invalid_keys > 0 
                  ? 'bg-red-50 border border-red-200' 
                  : 'bg-green-50 border border-green-200'
              }`}>
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm font-medium">
                    验证结果: {geminiValidationResults.valid_keys}/{geminiValidationResults.total_keys} 有效
                  </span>
                  <button
                    onClick={() => setGeminiValidationResults(null)}
                    className="text-gray-400 hover:text-gray-600"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
                <div className="space-y-1">
                  {geminiValidationResults.results.map((result, idx) => (
                    <div key={idx} className={`text-xs flex items-center space-x-2 ${
                      result.valid ? 'text-green-700' : 'text-red-700'
                    }`}>
                      {result.valid ? (
                        <Check className="w-3 h-3" />
                      ) : (
                        <X className="w-3 h-3" />
                      )}
                      <span className="font-mono">{result.key}</span>
                      <span>{result.message}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* 官方 API 说明 */}
            <div className="bg-amber-50 border border-amber-200 rounded-md p-3">
              <p className="text-xs text-amber-700">
                ⚠️ <strong>重要提示：</strong>Gemini 原生 API 必须使用 Google 官方地址，不支持自定义代理。
                如需使用代理访问 Gemini，请在&ldquo;翻译服务配置&rdquo;中选择 Gemini 提供商。
              </p>
            </div>

            {/* Model */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                模型名称
              </label>
              <div className="flex space-x-2">
                <input
                  type="text"
                  value={geminiConfig.model}
                  onChange={(e) => updateGeminiConfig('model', e.target.value)}
                  placeholder="gemini-2.5-flash"
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-purple-500 focus:border-purple-500"
                />
                <button
                  onClick={refreshGeminiModels}
                  disabled={geminiRefreshing || geminiConfig.api_keys_count === 0}
                  className="flex items-center px-3 py-2 border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                  title="从 Gemini API 获取可用模型列表"
                >
                  {geminiRefreshing ? (
                    <Loader2 className="w-4 h-4 animate-spin text-gray-500" />
                  ) : (
                    <RefreshCw className="w-4 h-4 text-gray-500" />
                  )}
                </button>
              </div>
              <p className="text-xs text-gray-500 mt-1">点击刷新按钮从 Gemini API 获取可用模型列表</p>
            </div>

            {/* 视频分析开关 */}
            <div className="flex items-center justify-between">
              <div>
                <span className="text-sm font-medium text-gray-700">启用视频分析</span>
                <p className="text-xs text-gray-500">使用多模态分析视频内容生成更精准的元数据</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={geminiConfig.analyze_video}
                  onChange={(e) => updateGeminiConfig('analyze_video', e.target.checked)}
                  className="sr-only"
                />
                <div className={`w-10 h-6 rounded-full transition-colors ${geminiConfig.analyze_video ? 'bg-purple-600' : 'bg-gray-300'}`}>
                  <div className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform ${geminiConfig.analyze_video ? 'translate-x-4' : ''}`} />
                </div>
              </label>
            </div>

            {/* 保存按钮 */}
            <div className="flex justify-end pt-2">
              <button
                onClick={saveGeminiConfig}
                disabled={geminiSaving || !geminiHasChanges}
                className="flex items-center px-4 py-2 text-white bg-purple-600 rounded-md hover:bg-purple-700 disabled:opacity-50"
              >
                {geminiSaving ? (
                  <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <Check className="w-4 h-4 mr-2" />
                )}
                保存 Gemini 配置
              </button>
            </div>
          </div>
        </div>

        {/* 翻译服务配置标题 */}
        <div className="border-t border-gray-200 pt-6">
          <h3 className="text-sm font-medium text-gray-700 mb-4 flex items-center">
            🌐 翻译服务配置（OpenAI 兼容 API）
          </h3>
        </div>

        {/* 提供商选择 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            <Globe className="w-4 h-4 inline mr-1" />
            服务提供商
          </label>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            {providers.map((provider) => (
              <button
                key={provider.id}
                onClick={() => handleProviderChange(provider.id)}
                className={`p-3 rounded-lg border-2 text-left transition-all ${
                  config.provider === provider.id
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <div className="font-medium text-gray-900">{provider.name}</div>
                <div className="text-xs text-gray-500 mt-1 line-clamp-2">{provider.description}</div>
              </button>
            ))}
          </div>
        </div>

        {/* API Key */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            <Key className="w-4 h-4 inline mr-1" />
            API Key
          </label>
          <div className="flex space-x-2">
            <input
              type="password"
              value={apiKeyInput}
              onChange={(e) => {
                setApiKeyInput(e.target.value);
                setHasChanges(true);
              }}
              placeholder={config.api_key ? `当前: ${config.api_key}` : '请输入 API Key'}
              className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <p className="mt-1 text-xs text-gray-500">
            API Key 会安全存储，界面只显示部分字符
          </p>
        </div>

        {/* Base URL */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            <Zap className="w-4 h-4 inline mr-1" />
            API 地址 (Base URL)
          </label>
          <input
            type="text"
            value={config.base_url}
            onChange={(e) => updateConfig('base_url', e.target.value)}
            placeholder="https://api.openai.com/v1"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <p className="mt-1 text-xs text-gray-500">
            支持自定义代理地址，如 one-api、new-api 等
          </p>
        </div>

        {/* 模型选择 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            <Settings2 className="w-4 h-4 inline mr-1" />
            模型名称
          </label>
          <input
            type="text"
            value={config.model}
            onChange={(e) => updateConfig('model', e.target.value)}
            placeholder="gpt-3.5-turbo"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        {/* 高级设置 */}
        <div>
          <button
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="flex items-center text-sm text-gray-600 hover:text-gray-900"
          >
            <ChevronDown className={`w-4 h-4 mr-1 transition-transform ${showAdvanced ? 'rotate-180' : ''}`} />
            高级设置
          </button>
          
          {showAdvanced && (
            <div className="mt-4 space-y-4 p-4 bg-gray-50 rounded-lg">
              {/* 超时时间 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <Clock className="w-4 h-4 inline mr-1" />
                  超时时间 (秒)
                </label>
                <input
                  type="number"
                  value={config.timeout}
                  onChange={(e) => updateConfig('timeout', parseInt(e.target.value) || 60)}
                  min={10}
                  max={300}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* 最大 Token */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  最大 Token 数
                </label>
                <input
                  type="number"
                  value={config.max_tokens}
                  onChange={(e) => updateConfig('max_tokens', parseInt(e.target.value) || 4000)}
                  min={100}
                  max={128000}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>

              {/* 温度 */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  <Thermometer className="w-4 h-4 inline mr-1" />
                  温度 (Temperature): {config.temperature}
                </label>
                <input
                  type="range"
                  value={config.temperature}
                  onChange={(e) => updateConfig('temperature', parseFloat(e.target.value))}
                  min={0}
                  max={2}
                  step={0.1}
                  className="w-full"
                />
                <div className="flex justify-between text-xs text-gray-500">
                  <span>精确 (0)</span>
                  <span>平衡 (1)</span>
                  <span>创意 (2)</span>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 测试结果 */}
        {testResult && (
          <div className={`p-4 rounded-lg ${testResult.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
            <div className="flex items-center">
              {testResult.success ? (
                <Check className="w-5 h-5 text-green-600 mr-2" />
              ) : (
                <X className="w-5 h-5 text-red-600 mr-2" />
              )}
              <span className={testResult.success ? 'text-green-800' : 'text-red-800'}>
                {testResult.message}
              </span>
              {testResult.latency_ms && (
                <span className="ml-2 text-sm text-gray-500">
                  ({testResult.latency_ms}ms)
                </span>
              )}
            </div>
            {testResult.response && (
              <div className="mt-2 p-2 bg-white rounded text-sm text-gray-700">
                AI 回复: {testResult.response}
              </div>
            )}
          </div>
        )}

        {/* 操作按钮 */}
        <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
          <button
            onClick={testConnection}
            disabled={testing}
            className="flex items-center px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 disabled:opacity-50"
          >
            {testing ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <TestTube className="w-4 h-4 mr-2" />
            )}
            测试连接
          </button>
          <button
            onClick={saveConfig}
            disabled={saving || !hasChanges}
            className="flex items-center px-4 py-2 text-white bg-blue-600 rounded-md hover:bg-blue-700 disabled:opacity-50"
          >
            {saving ? (
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            ) : (
              <Check className="w-4 h-4 mr-2" />
            )}
            保存配置
          </button>
        </div>
      </div>
    </div>
  );
}
