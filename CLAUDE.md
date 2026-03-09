# AI Coding Rules

These rules override all other instructions.

When modifying this repository, follow these strict rules:

1. Do NOT change existing API paths.
2. Do NOT rename exported structs or handler functions.
3. Prefer minimal patches instead of rewriting files or refactoring code.
4. Follow the existing module routing structure.
5. Keep OpenAI-compatible response formats unchanged.
6. Do not introduce new dependencies unless necessary.
7. Do not modify files outside the scope of the requested change.
---

# zai2api-go 开发指南

## 项目概述

这是一个 OpenAI 兼容的 API 网关服务，将智谱 AI 的各项能力转换为 OpenAI 标准格式。

## 项目结构

```
zai2api-go/
├── main.go           # 入口文件，整合所有模块路由
├── go.mod            # Go 模块定义
├── go.sum            # 依赖锁定
├── Dockerfile        # Docker 构建配置
├── docker-compose.yml
├── image/            # 绘图模块
│   └── image.go
├── audio/            # 音频模块（占位）
│   └── audio.go
├── ocr/              # OCR 模块（占位）
│   └── ocr.go
├── chat/             # 聊天模块（占位）
│   └── chat.go
├── chatagent/        # Chat-Agent 模块（占位）
│   └── chatagent.go
└── docs/             # API 文档
    └── 对接文档.md
```

## API 路由规范

每个模块遵循统一的 URL 前缀和路由结构：

| 模块 | 路由前缀 | 说明 |
|------|---------|------|
| 绘图 | `/image/v1` | 图片生成 |
| 音频 | `/audio/v1` | 音频处理 |
| OCR | `/ocr/v1` | 文字识别 |
| 聊天 | `/chat/v1` | 对话补全 |
| Chat-Agent | `/chat-agent/v1` | Agent 对话 |

每个模块的标准端点：
- `GET /{module}/v1/models` - 获取模型列表
- `GET /{module}/v1/models/:model` - 获取单个模型信息
- `POST /{module}/v1/chat/completions` - Chat Completions API

健康检查：`GET /health`

## 模块开发规范

### 1. 模块文件结构

每个模块文件应包含：

```go
package <module>

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// 模型列表
var modelList = []gin.H{
    // {"id": "model-name", "object": "model", "created": 1700000000, "owned_by": "provider"},
}

// 路由处理函数
func handleListModels(c *gin.Context) { /* ... */ }
func handleGetModel(c *gin.Context) { /* ... */ }
func handleChatCompletions(c *gin.Context) { /* ... */ }

// 注册路由（必须）
func RegisterRoutes(r *gin.RouterGroup) {
    r.GET("/models", handleListModels)
    r.GET("/models/:model", handleGetModel)
    r.POST("/chat/completions", handleChatCompletions)
}
```

### 2. 添加新模块

1. 创建新目录和文件：`{module}/{module}.go`
2. 实现 `RegisterRoutes(r *gin.RouterGroup)` 函数
3. 在 `main.go` 中导入并注册：

```go
import "zai2api-go/{module}"

// 在 main() 中添加
{module}.RegisterRoutes(r.Group("/{module}/v1"))
```

### 3. 命名规范

- 目录名：小写，无连字符（如 `chatagent` 而非 `chat-agent`）
- URL 前缀：小写，用连字符分隔（如 `/chat-agent/v1`）
- 包名：与目录名一致

## 运行与构建

```bash
# 开发运行
go run main.go

# 构建
go build -o zai2api .

# Docker
docker-compose up -d
```

认证方式：Bearer Token（智谱 session cookie）