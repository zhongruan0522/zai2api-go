# zai2api-go

OpenAI 兼容的智谱 AI API 网关，将智谱 AI 的各项能力转换为 OpenAI 标准格式。

## 功能

### Chat 对话
支持 GLM 系列模型的对话补全，包括 glm-5、glm-4.7v、glm-4-flash、glm-4-plus、glm-4v 等主流模型。

- SSE 流式响应
- reasoning_content 思考过程输出
- 多轮对话支持
- Function Calling 工具调用

### Image 绘图
基于智谱图像生成服务的 OpenAI 兼容接口。

- 支持 1K / 2K 分辨率
- 支持多种比例：1:1、3:4、4:3、16:9、9:16、21:9、9:21
- 提供 `/chat/completions` 和 `/images/generations` 两种调用方式

### 待实现模块
Audio 音频处理、OCR 文字识别、ChatAgent 智能体对话。

## 部署

### Docker（推荐）

```bash
docker-compose up -d
```

服务将在 8080 端口启动。

### 本地运行

需要 Go 1.21+ 环境：

```bash
go run main.go
```

## 认证

使用智谱 AI 的 session 作为 Bearer Token：

1. 登录 [chat.z.ai](https://chat.z.ai)
2. 从浏览器 Cookie 中获取 `session` 值
3. 请求时设置 `Authorization: Bearer <session>`

## 路由

| 模块 | 前缀 | 端点 |
|-----|------|-----|
| Chat | `/chat/v1` | `/models`, `/chat/completions` |
| Image | `/image/v1` | `/models`, `/chat/completions`, `/images/generations` |

## 环境变量

- `PORT`：服务端口，默认 8080

## License

MIT
