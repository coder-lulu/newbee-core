# 多租户架构方案设计 (Multi-Tenancy Architecture Design)

## 1. 概述 (Overview)

本文档旨在为 `newbee` 项目设计一个可扩展、安全且易于维护的多租户架构。该方案将作为核心平台的基础，并确保未来所有新建的微服务都能无缝集成多租户能力。

## 2. 方案选型 (Architecture Choice)

我们将采用 **共享数据库、共享 Schema、通过 `tenant_id` 进行数据隔离** 的方案。

- **数据库 (Database):** 单一数据库实例 (Shared Database)。
- **表结构 (Schema):** 所有租户共享同一套表结构 (Shared Schema)。
- **数据隔离 (Data Isolation):** 通过在每个需要隔离的表中增加 `tenant_id` 列来实现逻辑隔离。

### 2.1. 为什么选择此方案?

- **成本效益 (Cost-Effective):** 无需为每个租户创建独立的数据库或 Schema，显著降低了基础设施和维护成本。
- **高扩展性 (High Scalability):** 新增租户几乎没有额外成本，系统可以支持大量租户。新微服务可以轻松复用此模式。
- **易于维护和升级 (Easier Maintenance):** 所有租户共享代码库和数据库结构，升级和打补丁只需一次操作。
- **便于数据聚合 (Centralized Data):** 可以轻松地对所有租户的数据进行统一的分析、监控和报告。

### 2.2. 风险与挑战

- **数据安全风险:** 逻辑隔离的实现必须万无一失。任何代码或逻辑上的缺陷都可能导致租户数据泄露。
- **性能瓶颈:** "吵闹的邻居" (Noisy Neighbor) 问题，即某个大租户的高负载可能会影响其他租户的性能。
- **复杂性:** 需要在应用层面实现完善的数据隔离逻辑，对开发要求更高。

## 3. 核心设计 (Core Design)

### 3.1. `Tenant` 核心实体

系统将引入一个新的核心实体 `Tenant`。

- **路径:** `core/rpc/ent/schema/tenant.go`
- **核心字段:**
    - `id` (uint64): 租户的唯一标识符。
    - `name` (string): 租户的名称，用于显示。
    - `status` (enum): 租户状态 (e.g., `active`, `suspended`, `archived`)。
    - `created_at`, `updated_at`: 时间戳。

### 3.2. 数据表改造 (`ent` Schema)

所有需要按租户隔离的现有和未来实体（如 `User`, `Role`, `Api`, `Department` 等）都必须包含 `tenant_id` 字段。

为实现这一点，我们将创建一个 `TenantMixin`。

- **路径:** `core/rpc/ent/schema/mixin/tenant.go`
- **功能:**
    - 自动向 Schema 中添加 `tenant_id` (uint64) 字段。
    - 为 `tenant_id` 字段添加索引以优化查询性能。
    - 定义与 `Tenant` 实体的 `Edge` (关系)。

所有相关 Schema 文件都需要嵌入此 Mixin。

```go
// In user.go schema
func (User) Mixin() []ent.Mixin {
    return []ent.Mixin{
        // ... other mixins
        TenantMixin{},
    }
}
```

### 3.3. 上下文传递 (Context Propagation)

`tenant_id` 将作为请求上下文的核心部分，在服务的每一层之间传递。

1.  **用户登录 & 身份认证:**
    - 用户登录成功后，除了生成用户身份 token，还需确定其所属的 `tenant_id`。
    - 将 `tenant_id` 添加到 JWT (JSON Web Token) 的 `claims` 中。

2.  **API 网关层 (`core/api`):**
    - 修改 `core/api` 的 JWT 中间件。
    - 在解析并验证 JWT 后，从 `claims` 中提取 `user_id` 和 `tenant_id`。
    - 将 `tenant_id` 注入到 `context.Context` 中。

    ```go
    // Example in JWT middleware
    ctx = context.WithValue(r.Context(), "tenant_id", tenantID)
    ```

3.  **RPC 服务层 (`core/rpc`):**
    - `core/api` 通过 gRPC 调用 `core/rpc` 时，`context.Context` 会被自动传递。
    - `core/rpc` 的所有 `logic` 方法必须从 `context.Context` 中读取 `tenant_id` 以供后续的数据访问层使用。

### 3.4. 数据访问层隔离 (`ent` Hook)

这是确保数据安全隔离的核心机制。我们将利用 `ent` 的 `Hook` 功能在数据库操作层面强制执行租户隔离。

- **路径:** `core/rpc/ent/hook/tenant.go`
- **实现:**
    - 创建一个全局的 `Query` 钩子。
    - 该钩子会拦截所有 `Query` 类型的数据库操作。
    - **对于查询操作 (`SELECT`):**
        - 从 `context.Context` 中获取 `tenant_id`。
        - 如果 `tenant_id` 存在，则自动向查询的 `WHERE` 子句中添加 `AND tenant_id = ?` 条件。
        - 如果 `tenant_id` 不存在（例如，系统级操作），则跳过此逻辑。
    - **对于创建操作 (`INSERT`):**
        - 创建一个 `Mutation` 钩子。
        - 从 `context.Context` 中获取 `tenant_id`。
        - 自动将 `tenant_id` 填充到新创建的实体记录中。
    - **对于更新/删除操作 (`UPDATE`/`DELETE`):**
        - `Query` 钩子同样适用于 `UpdateOne`, `DeleteOne` 等操作，因为它们内部也依赖于查询来定位记录，从而确保只能修改/删除属于自己租户的数据。

```go
// Example of ent hook registration in client.go
func (c *Client) init() {
    c.Use(hook.Tenant()) // Register the tenant hook
}
```

## 4. 对未来微服务的影响 (Impact on Future Microservices)

此设计为未来的微服务化奠定了基础。任何新的微服务都必须遵循以下约定：

1.  **共享公共库:** 创建一个内部的 `go-commons` 或类似的基础库，包含：
    - `TenantMixin` 定义。
    - `ent` 的租户隔离钩子实现。
    - 用于解析 JWT 并提取租户信息的标准中间件。
2.  **服务开发流程:**
    - 在 `ent` schema 中使用 `TenantMixin`。
    - 在 `ent` 客户端初始化时注册租户钩子。
    - 在 `go-zero` 服务定义中使用标准的 JWT 中间件。
    - 业务逻辑代码无需（也不应该）直接处理 `tenant_id` 的过滤逻辑，应完全信赖底层钩子的隔离能力。

通过这种方式，我们可以确保所有服务在多租户方面行为一致，大大降低了新服务的开发和集成成本。
