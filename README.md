# NewBee Core

> 基于 simple-admin-core 二次开发的多租户核心服务，但已经发生深度架构重构，与上游项目不再兼容。请在新项目中全量采用本文档的指引。

## 项目概览

NewBee Core 是 NewBee 平台的后台核心服务，提供多租户账号体系、权限与数据治理、审计追踪、OAuth 与第三方任务调度等能力。项目由两个 Go 可执行组件组成：

- `api/`：基于 go-zero `rest` 的 HTTP 服务，负责开放接口、认证、审计与租户态上下文管理。
- `rpc/`：基于 go-zero `zrpc` 与 ent ORM 的 gRPC 服务，承载核心业务逻辑、RBAC 权限、数据权限控制与审计日志写入。

核心依赖统一抽象在 [`github.com/coder-lulu/newbee-common`](https://github.com/coder-lulu/newbee-common) 中，可复用跨服务的中间件、配置、国际化与 Casbin 适配器实现。

## 目录结构

```
api/            # HTTP 服务入口、路由、Handler 与配置
rpc/            # gRPC 服务、ent Schema、业务逻辑与插件
deploy/         # Docker Compose 与 Kubernetes 示例部署文件
docs/           # 核心治理指南（Casbin 迁移、初始化修复等）
Makefile        # 常用开发命令（构建、测试、代码生成）
go.mod          # 模块定义，使用 Go 1.25
```

更多细节：
- `api/internal/svc/service_context.go`：整合 Casbin、Redis、RPC 客户端与统一中间件链，并提供审计资源缓存、降级策略等扩展能力。
- `api/etc/core.yaml` & `rpc/etc/core.yaml`：默认配置模板，集中管理数据库、Redis、Casbin、统一中间件（认证、审计、租户校验、数据权限、权限校验、加密）等选项。
- `deploy/docker-compose/*`：覆盖一体化部署、核心服务独立部署、存储与消息依赖（MySQL、PostgreSQL、Redis、RocketMQ 等）的基础编排文件。

## 关键特性

- **统一中间件框架**：通过 `newbee-common/middleware/integration` 集成认证、租户校验、数据权限、审计、权限判定、加密响应等插件，并支持优雅关闭与健康检查降级。
- **审计增强**：异步审计写入、资源名缓存、真实客户端 IP 解析、响应体可选捕获，配置详见 `docs/COMMON_AUDIT_MIDDLEWARE_GUIDE.md`（位于上游 `common` 仓库）。
- **多租户与数据权限**：依托 Casbin + 自研规则引擎，支持跨租户 API 权限与数据范围控制，相关迁移说明在 `docs/CASBIN_MIGRATION_*.md` 中。
- **OAuth 与外部服务**：内建对 simple-admin-job、simple-admin-message-center、第三方 OAuth Provider 的客户端封装，可按需在配置中启用。
- **可观测性**：Prometheus 指标端点默认开启，支持 Zipkin/OTLP 链路追踪（可在配置中打开 `Telemetry` 栏位）。

## 环境要求

- Go 1.25 及以上（项目 go.mod 已指定 1.25.1）。
- MySQL / PostgreSQL / SQLServer（默认模板使用 MySQL），Redis 作为缓存与分布式锁。
- Protobuf / go-zero 开发工具链（`goctls`、`swagger`、`ent`）。
- 建议配套 NewBee 平台的 `common`、`nb-agent`、`ui` 子项目共同使用。

## 快速开始

1. **克隆仓库**（建议配合上层 `newbee` 单体仓库使用）。
2. **准备配置**：复制 `api/etc/core.yaml`、`rpc/etc/core.yaml`，按环境修改数据库、Redis、`middleware.audit.realIpHeader` 等参数。
3. **初始化数据库**：
   - 启动 RPC 服务后访问 `POST /core/init/database` 完成基础数据初始化。
   - 若需要同步 Job / 消息中心，请分别调用 `POST /core/init/job_database`、`POST /core/init/mcms_database`。
4. **启动服务**：
   ```bash
   go run ./rpc/core.go   # 启动 RPC
   go run ./api/core.go   # 启动 API
   ```
5. **验证健康状态**：访问 `/core/health`、`/metrics` 或使用 `curl` 检查主要接口。

## 常用命令

```bash
make fmt          # 格式化代码
make test         # 运行 API 与 RPC 单元测试
make lint         # 执行 golangci-lint（需先 make tools）
make gen-api      # 基于 goctl 生成 API 代码 & Swagger
make gen-rpc      # 基于 proto 生成 RPC 客户端/服务端
make gen-ent      # 基于 ent schema 生成 ORM 代码
make docker       # 构建 API/RPC Docker 镜像
```

> 温馨提示：go-zero 配置支持 `ENV_VAR=value` 方式覆盖，生产环境请通过环境变量或配置中心注入敏感信息。

## 配置要点

- `Middleware.audit`：必须根据部署环境设置 `realIpHeader`，并评估是否开启 `captureResponseBody`。
- `Middleware.permission` / `Middleware.dataPerm`：依赖 RPC Casbin 策略，禁用时会影响 RBAC、数据隔离能力。
- `CoreRpc` / `JobRpc` / `McmsRpc`：可按需启停，启用后需保证对应服务可用；API 服务内置健康降级防止雪崩。
- `ProjectConf`：管理注册、登录、验证码策略及默认租户实体。
- `Encryption`：响应加密器默认启用，务必替换生产密钥。

## 部署参考

- `deploy/docker-compose/all_in_one`：一次性拉起核心依赖，用于本地联调或 PoC。
- `deploy/docker-compose/core-only`：适合已有基础设施的环境，仅运行核心 API/RPC。
- `deploy/k8s`：提供 Kubernetes 部署样板，需结合实际集群调整 ConfigMap / Secret。
- `kill_core.sh`：快速释放调试端口（默认 9100/9101 系列）。

## 与 simple-admin-core 的差异

- 升级为 `newbee-common` 提供的统一中间件、配置、国际化与插件管理，替换了 simple-admin-core 原有的散装接入方式。
- 审计、权限、数据域、租户管理均经过重构，与上游路由、配置项、数据库初始化流程存在差异，不支持直接覆盖式升级。
- RPC 层新增插件体系、健康降级与缓存策略，多数逻辑文件经过重写；Ent Schema 也根据 NewBee 平台需求扩展了字段与 Hook。
- 配套 `docs/` 与 `deploy/` 目录提供新的运维指引，与 simple-admin-core 文档结构不兼容。

因此，请将 NewBee Core 视为全新品类服务，只保留 simple-admin-core 的基础理念和部分协议，而非可回滚的分支。

## 相关文档

- `docs/CASBIN_MIGRATION_GUIDE.md`：Casbin 迁移全流程指南。
- `docs/INIT_DATABASE_TABLE_FIX_REPORT.md`：初始化数据修复记录。
- `rpc/docs/TENANT_API_PERMISSION_INIT.md`：租户 API 权限初始化说明。
- 更多审计、中间件使用说明请参考 `newbee-common` 仓库中的《统一审计中间件指南》《统一数据权限中间件指南》。

## 许可证

项目遵循 Apache License 2.0，与 simple-admin-core 保持一致；若引入第三方资源，请遵循其各自许可证要求。

