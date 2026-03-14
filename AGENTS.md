# zai2api-go 项目说明

## 架构

```
zai2api-go/
├── backend/                # Go 后端 (Gin 框架)
│   ├── main.go             # 入口：只负责加载配置、初始化模块、启动 HTTP 服务
│   ├── config/             # 配置加载（环境变量）
│   ├── database/           # 数据库连接、迁移、初始化数据
│   ├── auth/               # JWT 登录/鉴权中间件
│   ├── models/             # GORM 数据模型（每个文件对应一张表或一组相关模型）
│   ├── handlers/           # HTTP 请求处理器（按功能域拆分文件，如 token.go、apikey.go、ocr.go）
│   ├── services/           # 业务逻辑层（如 token 选择器、渠道复用逻辑）
│   ├── common/             # 跨渠道通用逻辑（如定时任务调度器）
│   ├── router/             # 路由注册、CORS、静态文件服务
│   ├── ocr/                # OCR 渠道服务对接（上游请求、响应转换器、类型定义）
│   └── go.mod
├── frontend/               # Next.js 前端
│   ├── src/
│   │   ├── app/                 # App Router 页面
│   │   ├── components/ui/       # shadcn/ui 组件
│   │   └── lib/                 # 工具函数
│   └── package.json
├── Dockerfile              # 多架构 Docker 构建
├── docker-compose.yml      # 部署配置
└── .github/workflows/      # GitHub Actions CI/CD
```

## 技术栈

- **后端**: Go + Gin + CORS
- **前端**: Next.js (App Router) + TypeScript + Tailwind CSS v4 + shadcn/ui

## 开发命令

```bash
# 后端 (端口 8080)
cd backend && go run main.go

# 前端 (端口 3000)
cd frontend && npm run dev
```

## 部署

- 支持 Docker 多架构部署 (AMD64/ARM64)
- 推荐使用 `docker compose` 本地/单机部署（包含 Postgres）
- 首次部署：`cp .env.example .env` 后修改 `ADMIN_PASSWORD` / `JWT_SECRET` 再执行 `docker compose up -d`
- GitHub Actions 自动构建推送到 ghcr.io

## 环境变量与安全提示

- 生产环境必须设置强密码/强密钥：`ADMIN_PASSWORD`（>= 12 chars）、`JWT_SECRET`（>= 32 chars 随机字符串）
- 如部署在反向代理后，请设置 `TRUSTED_PROXIES`（否则 `ClientIP()` 可能不准确或可被伪造）
- 若需要从其他域访问后端（前后端分离），请设置 `CORS_ALLOW_ORIGINS`
- 详细审计与加固说明见：`SECURITY_AUDIT.md`

## 关于BUG反馈

1. 用户部署了一个基于最新的云端commit的项目实例，其位于http://45.205.31.20:10000/,账户admin，密码admintest，为测试专用平台
2. 可以使用浏览器MCP进行复测

## 代码规范

### 通用

- 本项目为前后端一体仓库，`backend/` 和 `frontend/` 同级存放
- 提交代码前确保 `cd backend && go build ./...` 和 Docker 构建均正常
- 前端 API 请求后端地址为 `http://localhost:8080`

### Go 后端模块化规则

- **main.go 只做三件事**：加载配置 → 初始化各模块 → 启动 HTTP 服务，不包含业务逻辑
- **按职责拆包，不按技术分层堆砌**：新增渠道（如 audio、chat）应各自建包（`audio/`、`chat/`），包内放上游对接、类型定义、响应转换器
- **跨渠道共用逻辑放 `common/`**：如定时调度器、通用工具函数
- **文件命名**：包内文件以功能命名（`converter.go`、`upstream.go`、`types.go`），避免 `utils.go`、`helper.go` 这类模糊命名
- **handlers/ 只放 HTTP 层**：请求解析、参数校验、响应返回；调用 services 和渠道包完成业务，不直接写上游请求和响应转换
- **models/ 按实体拆文件**：每个文件对应一张表或一组关联模型，文件名与模型名一致（`token.go`、`user.go`）
