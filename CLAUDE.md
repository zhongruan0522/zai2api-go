# zai2api-go 项目说明

## 架构

```
zai2api-go/
├── backend/          # Go 后端 (Gin 框架)
│   ├── main.go       # 入口文件
│   └── go.mod
├── frontend/         # Next.js 前端
│   ├── src/
│   │   ├── app/           # App Router 页面
│   │   ├── components/ui/ # shadcn/ui 组件
│   │   └── lib/           # 工具函数
│   └── package.json
├── Dockerfile            # 多架构 Docker 构建
├── docker-compose.yml    # 部署配置
└── .github/workflows/    # GitHub Actions CI/CD
```

## 技术栈

- **后端**: Go + Gin + CORS
- **前端**: Next.js 15 + TypeScript + Tailwind CSS v4 + shadcn/ui

## 开发命令

```bash
# 后端 (端口 8080)
cd backend && go run main.go

# 前端 (端口 3000)
cd frontend && npm run dev
```

## 部署

- 支持 Docker 多架构部署 (AMD64/ARM64)
- 使用 `docker-compose up -d` 本地部署
- GitHub Actions 自动构建推送到 ghcr.io

## 注意事项

- 修改代码时确保 Docker 构建正常
- 前端 API 请求后端地址为 `http://localhost:8080`
