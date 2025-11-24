# Gemini 多模态集成指南

## 概述

本项目已集成 Google Gemini 多模态 AI 服务，可以直接分析视频内容生成高质量的标题、描述和标签，无需依赖字幕。

## 功能特性

### 1. 多模态视频分析

- **直接分析视频**：Gemini 可以"观看"视频，理解画面内容、场景、人物动作等
- **音频理解**：识别背景音乐、语音内容和情绪
- **综合生成**：基于视觉和听觉信息生成更准确的元数据

### 2. 智能回退机制

- Gemini 视频分析（如果启用）
- Gemini 文本分析（字幕）
- DeepSeek 文本分析（默认/回退）

## 配置说明

### 获取 API 密钥

1. 访问 [Google AI Studio](https://aistudio.google.com/app/apikey)
2. 登录 Google 账号
3. 点击"Create API Key"创建密钥
4. 复制生成的 API 密钥

### 配置文件

在 `config.toml` 中添加以下配置：

```toml
[GeminiConfig]
  enabled = true                   # 启用Gemini服务
  api_key = "YOUR_API_KEY_HERE"    # 填入你的API密钥
  model = "gemini-1.5-pro"         # 推荐使用 gemini-1.5-pro
  timeout = 120                    # 超时时间（秒）
  max_tokens = 8000                # 最大输出token数
  use_for_metadata = true          # 使用Gemini生成元数据（优先于DeepSeek）
  analyze_video = true             # 启用视频分析（多模态）
  video_sample_frames = 0          # 0=上传完整视频
```

### 配置项说明

| 配置项                | 类型   | 默认值           | 说明                       |
| --------------------- | ------ | ---------------- | -------------------------- |
| `enabled`             | bool   | false            | 是否启用 Gemini 服务       |
| `api_key`             | string | ""               | Google AI API 密钥         |
| `model`               | string | "gemini-1.5-pro" | 使用的模型                 |
| `timeout`             | int    | 120              | API 调用超时时间（秒）     |
| `max_tokens`          | int    | 8000             | 最大输出 token 数          |
| `use_for_metadata`    | bool   | false            | 是否使用 Gemini 生成元数据 |
| `analyze_video`       | bool   | true             | 是否分析视频文件（多模态） |
| `video_sample_frames` | int    | 0                | 视频采样帧数（0=完整视频） |

## 使用场景

### 场景 1：纯视频分析（无字幕）

适用于：游戏集锦、风景视频、无声演示等

```toml
[GeminiConfig]
  enabled = true
  use_for_metadata = true
  analyze_video = true  # 启用视频分析
```

**优势**：

- 即使没有字幕，也能生成准确的标题和描述
- 可以识别游戏画面、场景、人物等视觉元素

### 场景 2：字幕+视频综合分析

适用于：有字幕的视频，需要更准确的元数据

```toml
[GeminiConfig]
  enabled = true
  use_for_metadata = true
  analyze_video = true
```

**优势**：

- 结合字幕和视觉信息，生成更全面的描述
- 可以捕捉字幕中未提及的视觉细节

### 场景 3：仅字幕分析（节省成本）

适用于：有完整字幕的视频，想节省 API 成本

```toml
[GeminiConfig]
  enabled = true
  use_for_metadata = true
  analyze_video = false  # 关闭视频分析
```

**优势**：

- 比视频分析更快、更便宜
- 仍然比 DeepSeek 更准确（Gemini 的文本理解能力更强）

## 成本对比

| 方案        | 输入成本 | 输出成本 | 适用场景             |
| ----------- | -------- | -------- | -------------------- |
| DeepSeek    | 极低     | 极低     | 有字幕的视频         |
| Gemini 文本 | 低       | 低       | 有字幕，需要更高质量 |
| Gemini 视频 | 中等     | 低       | 无字幕或需要视觉分析 |

**Gemini 1.5 Pro 定价**（参考）：

- 文本输入：$0.00125 / 1K tokens
- 视频输入：$0.00125 / 1K tokens（按视频时长计算）
- 输出：$0.005 / 1K tokens

## 工作流程

### 视频分析模式

```
1. 下载视频文件
2. 上传视频到 Gemini File API
3. 等待 Gemini 处理视频
4. 调用 Gemini 生成元数据
5. 保存标题、描述、标签
6. 上传到 Bilibili
```

### 文本分析模式

```
1. 下载并翻译字幕
2. 提取字幕文本
3. 调用 Gemini 生成元数据
4. 保存标题、描述、标签
5. 上传到 Bilibili
```

## 常见问题

### Q: Gemini 视频分析需要多长时间？

A: 取决于视频长度，通常：

- 5 分钟视频：约 30-60 秒
- 10 分钟视频：约 1-2 分钟
- 30 分钟视频：约 3-5 分钟

### Q: 支持哪些视频格式？

A: Gemini 支持常见的视频格式：

- MP4
- MOV
- AVI
- FLV
- MPG
- WEBM
- WMV
- 3GPP

### Q: 视频大小有限制吗？

A: 是的，Gemini File API 限制：

- 单个文件最大 2GB
- 视频最长 1 小时

### Q: 如何查看 Gemini 的使用情况？

A: 查看日志输出，会显示：

- 视频上传状态
- 处理进度
- 生成的元数据
- Token 使用情况（如果启用）

### Q: Gemini 失败了怎么办？

A: 系统有自动回退机制：

1. 如果视频分析失败，回退到字幕分析
2. 如果 Gemini 完全失败，回退到 DeepSeek
3. 如果 DeepSeek 也失败，使用默认值

## 最佳实践

### 1. 选择合适的模型

- **gemini-1.5-pro**：最强大，适合复杂视频
- **gemini-2.0-flash-exp**：更快更便宜，适合简单视频

### 2. 优化成本

- 对于有完整字幕的视频，设置 `analyze_video = false`
- 使用 `video_sample_frames` 采样视频帧（减少上传大小）
- 批量处理视频时，考虑使用 DeepSeek

### 3. 提高准确性

- 确保视频质量良好（清晰度、音质）
- 对于重要视频，启用视频分析
- 检查生成的元数据，必要时手动调整

## 示例输出

### 输入

- 视频：一个关于黑神话悟空的游戏实况
- 时长：10 分钟
- 无字幕

### Gemini 视频分析输出

**标题**：

```
黑神话悟空实况：挑战黄风大王BOSS战
```

**描述**：

```
这是一场激烈的黑神话悟空BOSS战实况！视频中玩家挑战黄风大王，
展示了多种战斗技巧和策略。从开场的小怪清理，到BOSS的各个阶段
应对，每一个细节都充满看点。特别是在BOSS第二阶段的完美闪避和
连招输出，展现了高超的操作水平。视频还包含了装备选择和技能搭配
的讲解，适合想要挑战这个BOSS的玩家参考。
```

**标签**：

```
["黑神话悟空", "游戏实况", "BOSS战", "黄风大王", "动作游戏"]
```

## 技术细节

### Gemini Client 实现

位置：`internal/chain_task/handlers/gemini_client.go`

主要方法：

- `NewGeminiClient()` - 创建客户端
- `UploadFile()` - 上传视频文件
- `WaitForFileProcessing()` - 等待处理完成
- `GenerateMetadataFromVideo()` - 从视频生成元数据
- `GenerateMetadataFromText()` - 从文本生成元数据

### 元数据生成流程

位置：`internal/chain_task/handlers/generate_metadata.go`

主要方法：

- `Execute()` - 主执行流程
- `executeWithGeminiVideo()` - 视频分析
- `executeWithGeminiText()` - 文本分析
- `executeWithDeepSeek()` - DeepSeek 回退
- `saveMetadataResults()` - 保存结果

## 更新日志

### v1.0.0 (2024-11-24)

- ✅ 集成 Gemini 1.5 Pro 多模态服务
- ✅ 支持视频文件直接分析
- ✅ 支持字幕文本分析
- ✅ 实现智能回退机制
- ✅ 添加配置选项和文档

## 参考链接

- [Google AI Studio](https://aistudio.google.com/)
- [Gemini API 文档](https://ai.google.dev/docs)
- [Gemini 定价](https://ai.google.dev/pricing)
- [支持的文件格式](https://ai.google.dev/gemini-api/docs/vision)
