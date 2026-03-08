# zai2api-go

OpenAI 兼容的智谱 AI API 网关，将智谱 AI 的各项能力转换为 OpenAI 标准格式。

## 功能特性

- OpenAI 兼容的 API 接口
- 支持 SSE 流式响应
- 支持 reasoning_content（思考过程）输出
- Docker 部署支持

## 已实现模块

| 模块 | 路由前缀 | 状态 | 说明 |
|------|---------|------|------|
| Chat | `/chat/v1` | ✅ 已完成 | 对话补全，支持 GLM 系列模型 |
| Image | `/image/v1` | ✅ 已完成 | 图片生成，支持多种分辨率和比例 |
| Audio | `/audio/v1` | 🚧 待实现 | 音频处理 |
| OCR | `/ocr/v1` | 🚧 待实现 | 文字识别 |
| ChatAgent | `/chat-agent/v1` | 🚧 待实现 | Agent 对话 |

## 支持的模型

### Chat 模块
- glm-5
- glm-4.7v
- glm-4-long
- glm-4-flash
- glm-4-plus
- glm-4-air
- glm-4-airx
- glm-4v
- glm-4v-plus

### Image 模块
- gemini-3-pro-image-1k / 2k
- 支持比例：1:1、3:4、4:3、16:9、9:16、21:9、9:21

## 快速开始

### 方式一：直接运行

```bash
# 克隆仓库
git clone https://github.com/GongYiDaLao/zai2api-go.git
cd zai2api-go

# 安装依赖
go mod tidy

# 运行
go run main.go
```

### 方式二：编译后运行

```bash
go build -o zai2api .
./zai2api
```

### 方式三：Docker 部署

```bash
docker-compose up -d
```

## API 使用

服务默认运行在 `http://localhost:8080`

### 健康检查

```bash
curl http://localhost:8080/health
```

### 获取模型列表

```bash
curl http://localhost:8080/chat/v1/models \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Chat Completions

```bash
curl http://localhost:8080/chat/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "glm-5",
    "messages": [{"role": "user", "content": "你好"}],
    "stream": true
  }'
```

### Image Generation

```bash
curl http://localhost:8080/image/v1/images/generations \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "model": "gemini-3-pro-image-1k-1-1",
    "prompt": "一只可爱的猫咪",
    "n": 1
  }'
```

## 认证说明

本服务需要智谱 AI 的 Token 进行认证：

1. 访问 [chat.z.ai](https://chat.z.ai) 并登录
2. 从浏览器 Cookie 中获取 `session` 值
3. 将该值作为 Bearer Token 使用

## 环境变量

| 变量 | 默认值 | 说明 |
|-----|-------|------|
| PORT | 8080 | 服务端口 |

## 项目结构

```
zai2api-go/
├── main.go           # 入口文件
├── chat/             # Chat 模块
│   └── chat.go
├── image/            # Image 模块
│   └── image.go
├── audio/            # Audio 模块（待实现）
├── ocr/              # OCR 模块（待实现）
├── chatagent/        # ChatAgent 模块（待实现）
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## 技术栈

- Go 1.21+
- Gin Web Framework
- OpenAI 兼容 API 格式

## License

MIT License
