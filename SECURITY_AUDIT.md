# 安全审计与加固记录（zai2api-go）

本文件用于记录一次面向公网部署场景的安全审计结论、已落地的修复，以及后续建议。

## 范围

- 后端：`backend/`（Gin + GORM + Postgres）
- 前端：`frontend/`（Next.js 管理界面）
- 部署：`Dockerfile` / `docker-compose.yml`

## 优先级说明

- P0：可被直接利用导致账号接管/数据破坏/高概率 DoS 的问题，优先修复
- P1：需要一定条件或影响面较大但仍建议尽快修复的问题
- P2：长期优化与工程化建议

## 已修复（按优先级提交）

### P0（commit: b76169a）

- 默认凭据/弱密钥上线防呆：生产模式（`GIN_MODE=release` 或 `APP_ENV=production`）下强制要求
  - `JWT_SECRET` 不可为默认值且长度 >= 32
  - `ADMIN_PASSWORD` 不可为默认值且长度 >= 12
  - 如需本地开发/临时环境，可设置 `ALLOW_INSECURE_DEFAULTS=true`
- 修复 GORM 条件注入风险：所有 `:id` 路由参数改为强制解析为 `uint` 后再用于查询/删除
  - 目标：阻断 `id=1 OR 1=1` / `id=1; ...` 一类字符串被当作 SQL 条件拼接的风险
- OCR 上传链路抗 DoS：上游转发改为流式 multipart（`io.Pipe`），避免把上传文件整体读入内存
- 对外路由请求体限制：按路由组增加 body size 限制
  - 管理端：`ADMIN_MAX_BODY_BYTES`（默认 2MB）
  - Image：`IMAGE_MAX_BODY_BYTES`（默认 1MB）
  - OCR：`OCR_MAX_BODY_BYTES`（默认 25MB）
- Gin 可信代理配置：默认不信任 `X-Forwarded-For`（`TRUSTED_PROXIES` 为空）
  - 若在 Nginx/Traefik 等反代后部署，需要显式设置 `TRUSTED_PROXIES`（IP/CIDR 列表）
- 生产默认关闭 CORS：仅在非生产环境默认允许 `http://localhost:3000,http://127.0.0.1:3000`
  - 生产如需跨域访问，请设置 `CORS_ALLOW_ORIGINS`
- 静态前端路由兜底加固：修复 `filepath.Join(frontendDir, path)` 里 path 以 `/` 开头导致的路径处理问题，并增加候选文件解析（`index.html` / `xxx/index.html` / `xxx.html`）
- HTTP Server 增加基础超时：替换 `r.Run`，添加 `ReadHeaderTimeout/ReadTimeout/WriteTimeout/IdleTimeout/MaxHeaderBytes`
- `docker-compose.yml` 生产配置改为强制提供 `ADMIN_USERNAME/ADMIN_PASSWORD/JWT_SECRET`

### P1（commit: cdfb197）

- 计数字段防溢出：将 Token 计数/限额字段从 `int` 调整为 `int64`
  - `total_call_count` / `daily_call_count` / `daily_limit` 迁移为 Postgres `bigint`
  - 启动时执行 `migrateCounterColumnTypes()` 做列类型加宽（注意：ALTER TABLE 可能短时锁表）
- 日志查询防止一次性拉爆：`GET /api/logs` 的 `page_size` 增加上限（最大 200）
- 日志写入稳定性：写入 `request_log` 前对 `error_code`（20）与 `error_msg`（500）做截断，避免超过 varchar 长度导致插入失败

## 新增/变更的环境变量

- 必须（生产）：
  - `ADMIN_USERNAME`
  - `ADMIN_PASSWORD`
  - `JWT_SECRET`
- 可选：
  - `ALLOW_INSECURE_DEFAULTS`：`true/false`，允许生产模式下继续使用默认弱配置（不建议）
  - `APP_ENV`：设置为 `production` 视为生产模式
  - `CORS_ALLOW_ORIGINS`：逗号分隔；支持 `*`（不建议在管理端开启 `*`）
  - `TRUSTED_PROXIES`：逗号分隔 IP/CIDR，用于正确解析 `ClientIP()`
  - `ADMIN_MAX_BODY_BYTES` / `IMAGE_MAX_BODY_BYTES` / `OCR_MAX_BODY_BYTES`
  - `UPSTREAM_MAX_RESPONSE_BYTES`：限制上游响应读取大小（默认 10MB）

## 仍建议推进的优化（未在本轮提交中落地）

### 路由与鉴权

- 公共 API 增加限流/配额：按 APIKey + 源 IP 双维度限流，防止被刷量导致带宽/DB/上游额度被耗尽
- 登录接口防爆破：对 `/api/login` 增加失败次数惩罚（临时锁定/退避）与审计日志
- 统一 APIKey 校验为中间件：减少重复代码、降低漏校验风险

### 前端与会话

- JWT 存储在 `localStorage` 存在 XSS 放大效应；建议改为 HttpOnly Cookie + CSRF 防护
- 管理端可增加二次确认/操作审计（批量删除/批量禁用等）

### 数据库与数据安全

- APIKey/上游 Token 明文存储：建议
  - APIKey：只存 hash（创建时展示一次，之后只展示前缀/末尾）
  - 上游 Token：应用层加密（AES-GCM + 环境密钥/KMS）
- `request_log` 增加保留期/分区：防止长期运行被刷量打满磁盘
- 生产环境启用 DB TLS（避免 `sslmode=disable`）并使用最小权限账户

### 可观测与工程化

- 增加安全相关告警：默认凭据使用、登录失败爆发、请求体过大、上游失败率飙升
- CI 增加依赖漏洞扫描：Go `govulncheck` / `gosec`，前端 `npm audit`

## 验证方式（开发机）

后端：

```bash
cd backend
go test ./...
```

## 备注

如果你在反向代理（Nginx/Traefik/Caddy）后部署，请务必正确配置 `TRUSTED_PROXIES`，否则 `ClientIP()` 可能只看到代理地址或被伪造。
