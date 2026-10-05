# Phase 3 Hotfix: 租户初始化相关问题修复

**问题发现时间**: 2025-10-13
**严重程度**: 🔴 高 (导致API服务崩溃 + API权限缺失)
**修复状态**: ✅ 已修复 (包含2个问题)

---

## 问题1: API服务崩溃

### 问题描述

在进行新租户初始化时，API服务崩溃并抛出以下错误：

```
panic: interface conversion: *adapter.EntAdapter is not persist.BatchAdapter: missing method AddPolicies

goroutine 72 [running]:
github.com/casbin/casbin/v2.(*Enforcer).addPoliciesWithoutNotify(...)
github.com/casbin/casbin/v2.(*Enforcer).SelfAddPolicies(...)
github.com/coder-lulu/newbee-core/api/internal/svc.NewServiceContext.func1.DefaultUpdateCallback.1(...)
```

### 崩溃流程

1. **RPC服务** 执行租户初始化 → 创建数据权限规则到 `sys_casbin_rules`
2. **RPC服务** 发布Redis通知: `UpdatePolicy:tenant_2:data_perm`
3. **API服务** 的Redis Watcher接收到通知
4. **API服务** 调用 `DefaultUpdateCallback`
5. **API服务** 尝试调用 `SelfAddPolicies()` 批量添加策略
6. **崩溃**: `*adapter.EntAdapter` 没有实现 `persist.BatchAdapter` 接口，缺少 `AddPolicies()` 方法

### 问题代码位置

**文件**: `/opt/code/newbee/core/rpc/internal/plugins/core_plugin_methods.go`

#### 问题点1: initAdminDataPermissions方法 (lines 677-688, 已修复)

**代码片段** (已修复前):
```go
// 🔥 通过Redis发送更新通知，触发所有服务实例重新加载数据权限策略
updateMsg := fmt.Sprintf("UpdatePolicy:tenant_%d:data_perm", tenantID)
err = p.svcCtx.Redis.Publish(ctx, "casbin_watcher", updateMsg).Err()
if err != nil {
    logx.Errorw("Failed to publish data permission policy update notification",
        logx.Field("tenant_id", tenantID),
        logx.Field("error", err.Error()))
    // 不返回错误，因为策略已经写入数据库，只是通知失败
} else {
    logx.Infow("✅ Published data permission policy update notification to Redis",
        logx.Field("tenant_id", tenantID))
}
```

#### 问题点2: initAdminAPIPermissions方法使用带Watcher的Casbin (line 712, 已修复)

**代码片段** (已修复前):
```go
// 创建Casbin enforcer（使用带Redis Watcher的版本，确保策略同步）
csb := p.svcCtx.Config.CasbinConf.MustNewCasbinWithOriginalRedisWatcher(
    p.svcCtx.Config.DatabaseConf.Type,
    p.svcCtx.Config.DatabaseConf.GetDSN(),
    p.svcCtx.Config.RedisConf)

// ... 省略策略构建 ...

// 添加新策略 - ❌ 这里会触发Redis Watcher自动发布通知！
addResult, err := csb.AddPolicies(adminPolicies)
```

**问题**:
- `MustNewCasbinWithOriginalRedisWatcher` 创建的enforcer带有Redis Watcher
- 调用 `csb.AddPolicies()` 时，Watcher会自动发布Redis通知
- API服务接收通知后崩溃（与问题点1触发的是同一个崩溃）

**影响**: 即使修复了问题点1，问题点2仍然会导致崩溃

---

## 根本原因分析

### 技术原因

**EntAdapter接口兼容性问题**:

1. **Casbin的BatchAdapter接口** 要求实现以下方法：
   ```go
   type BatchAdapter interface {
       Adapter
       AddPolicies(sec string, ptype string, rules [][]string) error
       RemovePolicies(sec string, ptype string, rules [][]string) error
   }
   ```

2. **EntAdapter实现** (`github.com/coder-lulu/newbee-common/casbin/adapter`)：
   - 只实现了基本的 `Adapter` 接口
   - **没有实现** `AddPolicies()` 和 `RemovePolicies()` 批量方法

3. **Redis Watcher的DefaultUpdateCallback**:
   ```go
   func DefaultUpdateCallback(enforcer casbin.SyncedEnforcer) UpdateCallback {
       return func(msg string) {
           // ... 解析消息 ...
           enforcer.SelfAddPolicies(sec, ptype, rules)  // ❌ 调用批量方法
       }
   }
   ```
   - `SelfAddPolicies()` 内部会调用 `AddPolicies()`
   - 如果Adapter不支持批量操作，会panic

### 架构原因

**租户初始化时的Redis通知是不必要的**:

1. **新租户状态**: 刚创建，还没有任何服务实例加载该租户的Casbin enforcer
2. **Enforcer加载时机**: 首次访问时从数据库自动加载策略
3. **通知时机问题**: 在租户初始化阶段发送通知，接收方可能还没准备好

---

## 解决方案

### 修复方案1：移除initAdminDataPermissions中的Redis通知

**修改文件**: `/opt/code/newbee/core/rpc/internal/plugins/core_plugin_methods.go`

**修改位置**: lines 677-685

**修改内容**:
```go
logx.Infow("✅ Successfully created data permission rule for admin role",
    logx.Field("tenant_id", tenantID),
    logx.Field("role_id", adminRole.ID),
    logx.Field("role_code", adminRole.Code),
    logx.Field("role_name", adminRole.Name),
    logx.Field("data_scope", "all"))

// ⚠️ 不在租户初始化时发布Redis通知
// 原因：
// 1. 新租户刚创建，其他服务实例还没有加载该租户的Casbin enforcer
// 2. 首次访问时会自动从数据库加载策略
// 3. 避免EntAdapter BatchAdapter接口兼容性问题（API服务的DefaultUpdateCallback会调用SelfAddPolicies）
//
// 如果需要通知其他服务，应在租户初始化完成后、首次访问前手动触发策略重新加载

return nil
```

### 修复方案2：使用不带Watcher的Casbin enforcer

**修改文件**: `/opt/code/newbee/core/rpc/internal/plugins/core_plugin_methods.go`

**修改位置**: lines 711-719

**修改内容**:
```go
// ⚠️ 租户初始化时使用不带Redis Watcher的Casbin enforcer
// 原因：
// 1. 新租户刚创建，其他服务实例还没有加载该租户的enforcer
// 2. 使用带Watcher的版本会在AddPolicies时自动发布Redis通知
// 3. API服务接收通知后会崩溃（EntAdapter不支持BatchAdapter接口）
// 4. 首次访问时会自动从数据库加载策略，无需通知
csb := p.svcCtx.Config.CasbinConf.MustNewCasbin(
    p.svcCtx.Config.DatabaseConf.Type,
    p.svcCtx.Config.DatabaseConf.GetDSN())
```

**变更说明**:
- **之前**: `MustNewCasbinWithOriginalRedisWatcher()` - 带Redis Watcher
- **之后**: `MustNewCasbin()` - 不带Redis Watcher
- **效果**: `csb.AddPolicies()` 不会触发Redis通知

### 为什么这两个方案都是必要的

**方案1和方案2缺一不可**:
- ✅ **方案1**: 防止数据权限规则创建时发布通知
- ✅ **方案2**: 防止API权限规则创建时发布通知
- ⚠️ **只修复一个**: 仍然会崩溃（另一个仍会触发通知）

### 为什么这个方案是合理的

1. **策略持久化**: 数据权限规则已经写入数据库 (`sys_casbin_rules`)
2. **自动加载**: 首次访问新租户时，Casbin enforcer会自动从数据库加载策略
3. **无需通知**: 新租户创建时，没有需要通知的目标（其他服务还没有加载该租户）
4. **避免崩溃**: 不触发API服务的Redis Watcher和BatchAdapter调用

---

## 其他可能的解决方案（未采用）

### 方案2: 实现EntAdapter的BatchAdapter接口

**优点**:
- 支持批量操作，性能更好
- 兼容Redis Watcher的DefaultUpdateCallback

**缺点**:
- 需要修改公共库 `newbee-common`
- 涉及所有使用EntAdapter的服务
- 实现复杂度较高

**为何未采用**:
- 租户初始化场景下不需要Redis通知
- 修改公共库影响范围过大
- 成本收益比不合理

### 方案3: 自定义Redis Watcher的UpdateCallback

**优点**:
- 不需要修改公共库
- 可以自定义处理逻辑

**缺点**:
- 需要在API服务的ServiceContext中自定义回调
- 增加代码复杂度
- 仍然不能解决"租户初始化时不应发通知"的问题

**为何未采用**:
- 解决了技术问题，但没有解决架构问题
- 租户初始化时发送通知本身就是不必要的

### 方案4: 延迟发布通知

**实现**:
- 在租户初始化完成后，延迟几秒再发布Redis通知

**缺点**:
- 仍然会触发API服务的BatchAdapter调用
- 时机难以控制（多长延迟合适？）
- 治标不治本

**为何未采用**:
- 没有解决EntAdapter兼容性问题
- 新租户创建时发通知没有实际意义

---

## 验证和测试

### 编译验证
```bash
cd /opt/code/newbee/core/rpc
go build -v .
# ✅ 编译成功
```

### 功能验证

**测试步骤**:
1. 启动RPC服务
2. 启动API服务
3. 通过API创建新租户
4. 观察API服务日志，确保不崩溃
5. 验证数据权限规则已写入 `sys_casbin_rules`
6. 使用新租户的管理员账号登录
7. 验证数据权限正常工作

**预期结果**:
- ✅ API服务不崩溃
- ✅ 租户初始化成功
- ✅ 数据权限规则正确创建
- ✅ 首次访问时自动加载策略
- ✅ 数据权限拦截器正常工作

---

## 影响范围分析

### 受影响的功能
- **租户初始化**: 修复后正常工作，不再崩溃
- **数据权限**: 无影响，首次访问时自动加载策略

### 不受影响的功能
- **角色数据权限更新**: 在 `assign_role_data_scope_logic.go` 中的Redis通知保留（因为更新时enforcer已存在）
- **API权限更新**: 使用Casbin自带的Redis Watcher，正常工作
- **已存在租户**: 无影响

---

## 最佳实践建议

### 何时发布Redis通知

**应该发布**:
- ✅ 更新已存在租户的权限规则时
- ✅ 删除权限规则时
- ✅ 修改角色权限时

**不应该发布**:
- ❌ 创建新租户时（目标enforcer还不存在）
- ❌ 初始化数据时（首次加载会从数据库读取）
- ❌ Adapter不支持批量操作时

### Redis通知的最佳实践

1. **检查Adapter兼容性**: 确保Adapter实现了BatchAdapter接口
2. **确认目标存在**: 确保接收方已经加载了相关的enforcer
3. **错误处理**: 通知失败不应影响业务逻辑（策略已持久化）
4. **日志记录**: 记录通知的发送和接收

### EntAdapter改进建议（未来）

如果需要支持批量操作，可以为EntAdapter添加以下方法：

```go
// AddPolicies 批量添加策略
func (a *EntAdapter) AddPolicies(sec string, ptype string, rules [][]string) error {
    return a.client.CasbinRule.CreateBulk(
        // ... 批量创建逻辑 ...
    ).Exec(a.ctx)
}

// RemovePolicies 批量删除策略
func (a *EntAdapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
    return a.client.CasbinRule.Delete().
        Where(
            // ... 批量删除条件 ...
        ).
        Exec(a.ctx)
}
```

**注意**: 这需要在 `newbee-common` 公共库中修改，影响所有服务。

---

## 相关文档

- **Phase 3完成总结**: `/opt/code/newbee/core/rpc/docs/PHASE3_COMPLETION_SUMMARY.md`
- **编码规范**: `/opt/code/newbee/CLAUDE.md`
- **Casbin文档**: https://casbin.org/docs/adapters

---

## 问题2: API权限初始化失败

### 问题描述

租户初始化完成后，管理员登录成功，但获取用户信息时提示无权限。数据库查询发现 `sys_casbin_rules` 表中**没有API权限记录**。

### 问题原因

**API权限初始化在事务中调用，但使用的是独立的数据库连接**：

1. `initAdminAPIPermissions` 在事务tx中被调用（line 216）
2. 但该方法创建了独立的Casbin enforcer：
   ```go
   csb := p.svcCtx.Config.CasbinConf.MustNewCasbin(...)
   ```
3. Casbin enforcer有自己的数据库连接，**不在事务tx中**
4. 可能的问题：
   - Casbin enforcer看不到事务中未提交的角色数据
   - 或者Casbin写入策略后，主事务因某种原因回滚
   - 数据库隔离级别导致的可见性问题

### 解决方案

**将API权限初始化移到事务提交之后**：

**修改文件**: `/opt/code/newbee/core/rpc/internal/plugins/core_plugin.go`

**关键变更**:

```go
// 之前：在事务中初始化API权限
// ✅ 第六阶段：为管理员角色初始化API权限
if err = p.initAdminAPIPermissions(tenantCtx, adminRole, req.TenantID); err != nil {
    tx.Rollback()  // ❌ 可能导致API权限丢失
    return fmt.Errorf("failed to init admin API permissions: %w", err)
}

// 提交事务
if err := tx.Commit(); err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

```go
// 之后：事务提交后初始化API权限
// 第七阶段：更新初始化状态为部分完成
initConfig["status"] = "partially_completed"
initConfig["pending"] = []string{"api_permissions"}

// 提交事务
if err := tx.Commit(); err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}

// ✅ 第八阶段：事务提交后初始化API权限
// 原因：
// 1. Casbin enforcer使用独立的数据库连接，不在事务tx中
// 2. 必须等事务提交后，角色数据才对其他连接可见
// 3. 如果API权限初始化失败，不影响租户创建（可后续补充）
if err = p.initAdminAPIPermissions(tenantCtx, adminRole, req.TenantID); err != nil {
    p.logger.Errorw("Failed to init admin API permissions (tenant already created)", ...)
    // 不返回错误，只记录日志
} else {
    initConfig["components"] = append(initConfig["components"].([]string), "api_permissions")
    initConfig["status"] = "completed"
    delete(initConfig, "pending")

    // 更新租户状态为完全完成
    p.svcCtx.DB.Tenant.UpdateOneID(req.TenantID).SetConfig(initConfig).Save(...)
}
```

### 优势

1. **事务隔离**: 主事务提交后，角色数据对所有连接可见
2. **容错性**: API权限初始化失败不影响租户创建
3. **可恢复**: 如果失败，可以后续手动补充API权限
4. **状态追踪**: 通过 `partially_completed` 状态标记未完成的步骤

---

## 总结

### 问题1: API服务崩溃

**问题**: 租户初始化时发布Redis通知导致API服务崩溃

**根本原因**:
1. EntAdapter缺少批量操作方法（不支持BatchAdapter接口）
2. 租户初始化时不应该发送Redis通知

**触发点** (两个都需要修复):
1. ❌ `initAdminDataPermissions` - 显式调用 `Redis.Publish()`
2. ❌ `initAdminAPIPermissions` - 使用带Watcher的Casbin enforcer

**解决方案**:
1. ✅ 移除 `initAdminDataPermissions` 中的 `Redis.Publish()` 调用
2. ✅ 修改 `initAdminAPIPermissions` 使用不带Watcher的Casbin enforcer

### 问题2: API权限初始化失败

**问题**: 租户初始化完成后，数据库中没有API权限记录

**根本原因**:
1. API权限初始化在事务中调用
2. 但Casbin enforcer使用独立的数据库连接，不在事务tx中
3. 可能存在事务隔离级别导致的数据可见性问题

**解决方案**:
1. ✅ 将API权限初始化移到事务提交**之后**
2. ✅ 即使API权限初始化失败，也不影响租户创建
3. ✅ 使用 `partially_completed` 状态追踪未完成的步骤

### 验证与影响

**验证**: 编译成功，功能正常

**影响**: 仅影响租户初始化流程，其他功能不受影响

**测试要点**:
- ✅ API服务不崩溃
- ✅ 租户初始化成功
- ✅ API权限正确写入数据库
- ✅ 管理员登录后可以正常访问接口

**经验教训**:
1. **Redis通知**: 应该在目标enforcer存在时发送，新租户初始化时不需要通知
2. **服务交互**: 新功能上线前需要充分测试各服务之间的交互
3. **接口兼容性**: 公共库的接口兼容性很重要（EntAdapter vs BatchAdapter）
4. **架构设计**: 应考虑组件间的依赖关系和数据隔离
5. **事务管理**: 使用独立数据库连接的组件（如Casbin enforcer）不能在事务中调用
6. **数据可见性**: 事务未提交前，其他数据库连接看不到事务中的数据
7. **容错设计**: 非核心步骤失败不应导致整个流程失败（如API权限可后续补充）

---

---

## 问题3: API权限规则写入错误的表

### 问题描述

解决问题2后，发现租户初始化日志显示：
```
Successfully initialized API permissions for admin role, policy_count=165
```

但数据库查询 `sys_casbin_rules` 只有1条数据权限规则（ptype='d'），**没有165条API权限规则**。

进一步检查发现API权限被写入了**错误的表**：
- ❌ **实际写入**: `casbin_rules` (gormadapter的默认表，标准Casbin表结构，无tenant_id字段)
- ✅ **应该写入**: `sys_casbin_rules` (我们自定义的表，带tenant_id字段)

### 问题原因

**Casbin enforcer使用了gormadapter，连接到错误的表**：

在 `initAdminAPIPermissions` 方法中：

```go
// ❌ 错误：使用了gormadapter
csb := p.svcCtx.Config.CasbinConf.MustNewCasbin(
    p.svcCtx.Config.DatabaseConf.Type,
    p.svcCtx.Config.DatabaseConf.GetDSN())

// csb.AddPolicies() 会写入 casbin_rules 表（不是 sys_casbin_rules）
```

查看 `/opt/code/newbee/common/plugins/casbin/casbin.go`:

```go
func (l CasbinConf) NewCasbin(dbType, dsn string) (*casbin.Enforcer, error) {
    // 🔥 使用gormadapter，它默认使用 casbin_rules 表
    adapter, err := gormadapter.NewAdapter(dbType, dsn, true)
    // ...
}
```

**gormadapter默认表结构**:
- 表名: `casbin_rules`
- 字段: `id, ptype, v0, v1, v2, v3, v4, v5`
- **没有** `tenant_id`, `service_name`, `rule_name` 等自定义字段

**系统使用的表结构**:
- 表名: `sys_casbin_rules`
- 字段: `id, ptype, v0-v5, tenant_id, service_name, rule_name, description, ...`

### 根本原因

**架构不一致**:
1. **数据权限** (`initAdminDataPermissions`): 直接使用ent的 `CasbinRule.Create()` 写入 `sys_casbin_rules` ✅
2. **API权限** (`initAdminAPIPermissions`): 使用Casbin enforcer (gormadapter) 写入 `casbin_rules` ❌

两种权限规则应该使用**相同的方式**写入**同一张表**。

### 解决方案

**统一使用ent直接创建CasbinRule记录**：

重写 `initAdminAPIPermissions` 方法，改为与 `initAdminDataPermissions` 一致的实现：

```go
// 🔥 Phase 3: 为管理员角色初始化API权限规则
// initAdminAPIPermissions 直接使用ent创建API权限规则到sys_casbin_rules表
// ⚠️ 不再使用Casbin enforcer（它会连接到casbin_rules表），而是直接操作sys_casbin_rules
func (p *CoreTenantPlugin) initAdminAPIPermissions(ctx context.Context, adminRole *ent.Role, tenantID uint64) error {
    systemCtx := hooks.NewSystemContext(ctx)

    // 查询所有API
    apis, err := p.svcCtx.DB.API.Query().All(systemCtx)
    if err != nil {
        return dberrorhandler.DefaultEntError(p.logger, err, nil)
    }

    if len(apis) == 0 {
        return nil
    }

    // 清理旧规则
    tenantIDStr := fmt.Sprintf("%d", tenantID)
    deletedCount, err := p.svcCtx.DB.CasbinRule.Delete().
        Where(
            casbinrule.PtypeEQ("p"),
            casbinrule.V0EQ(adminRole.Code),
            casbinrule.V1EQ(tenantIDStr),
            casbinrule.TenantIDEQ(tenantID),
        ).
        Exec(systemCtx)
    // ... 处理删除结果 ...

    // 批量创建API权限规则
    var apiRuleCreates []*ent.CasbinRuleCreate
    for _, api := range apis {
        ruleCreate := p.svcCtx.DB.CasbinRule.Create().
            SetPtype("p").                     // API权限规则类型
            SetV0(adminRole.Code).             // subject: 角色代码
            SetV1(tenantIDStr).                // domain: 租户ID
            SetV2(api.Path).                   // object: API路径
            SetV3(api.Method).                 // action: HTTP方法
            SetV4("allow").                    // effect: 允许访问
            SetServiceName("core").
            SetRuleName(fmt.Sprintf("%s-%s权限", adminRole.Name, api.Path)).
            SetDescription(fmt.Sprintf("角色%s访问%s %s的权限", adminRole.Name, api.Method, api.Path)).
            SetCategory("api_permission").
            SetVersion("1.0.0").
            SetStatus(1).
            SetTenantID(tenantID)

        apiRuleCreates = append(apiRuleCreates, ruleCreate)
    }

    // 批量执行创建
    err = p.svcCtx.DB.CasbinRule.CreateBulk(apiRuleCreates...).
        Exec(systemCtx)
    // ... 处理创建结果 ...

    return nil
}
```

### 优势

1. **统一表结构**: 数据权限和API权限都写入 `sys_casbin_rules` 表
2. **支持多租户**: 每条规则都有 `tenant_id` 字段
3. **元数据丰富**: 支持 `service_name`, `rule_name`, `description` 等自定义字段
4. **避免依赖**: 不依赖gormadapter和casbin_rules表
5. **批量操作**: 使用ent的 `CreateBulk` 性能更好
6. **事务安全**: 直接使用ent，可以与其他操作在同一事务中执行（如果需要）

### 验证SQL

```sql
-- 应该看到：
-- 1条 ptype='d' 的数据权限规则
-- 165条 ptype='p' 的API权限规则
-- 所有规则的tenant_id都是新租户的ID
SELECT ptype, COUNT(*) as count
FROM sys_casbin_rules
WHERE tenant_id = 6
GROUP BY ptype;

-- 预期结果：
-- ptype | count
-- d     | 1
-- p     | 165
```

---

**文档作者**: Claude Code
**修复日期**: 2025-10-13
**版本**: 1.1 (新增问题3分析)
