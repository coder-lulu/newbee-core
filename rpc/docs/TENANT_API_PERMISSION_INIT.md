# 租户API权限初始化功能说明

## 📋 概述

本文档说明了在新租户初始化过程中，如何自动为管理员角色分配全部API接口权限。

## 🎯 功能目标

**问题**：新租户初始化时，管理员角色虽然已创建，但没有分配任何API接口权限，导致管理员无法调用任何接口。

**解决方案**：在租户初始化流程中增加第六阶段 - API权限初始化，自动为新租户的管理员角色分配全部API接口权限。

## 🔧 实现方案

### 1. 修改文件

#### `/opt/code/newbee/core/rpc/internal/plugins/core_plugin.go`

**修改位置**：`executeCoreTenantInit` 方法

**变更内容**：在创建管理员角色和用户后，增加API权限初始化阶段

```go
// 第五阶段：创建管理员角色和用户
adminRole, adminUser, err := p.initAdminRoleAndUser(tenantCtx, tx, req, dept)
if err != nil {
    tx.Rollback()
    return fmt.Errorf("failed to init admin role and user: %w", err)
}
initConfig["components"] = append(initConfig["components"].([]string), "admin_role", "admin_user")

// ✅ 第六阶段：为管理员角色初始化API权限
if err = p.initAdminAPIPermissions(tenantCtx, adminRole, req.TenantID); err != nil {
    tx.Rollback()
    return fmt.Errorf("failed to init admin API permissions: %w", err)
}
initConfig["components"] = append(initConfig["components"].([]string), "api_permissions")

// 第七阶段：完成初始化状态更新
```

#### `/opt/code/newbee/core/rpc/internal/plugins/core_plugin_methods.go`

**新增方法**：`initAdminAPIPermissions`

**功能说明**：
1. 使用系统上下文查询所有API（因为API表是系统级数据）
2. 为管理员角色构建Casbin权限策略
3. 清除旧策略（如果存在）
4. 添加新策略

### 2. 权限策略格式

使用 **RBAC with Domains** 模型：

```
[sub, domain, obj, act, eft]
```

| 字段 | 说明 | 示例 |
|------|------|------|
| sub | 角色代码 | "admin" |
| domain | 租户ID | "2" |
| obj | API路径 | "/user/list" |
| act | HTTP方法 | "POST" |
| eft | 效果 | "allow" |

**示例策略**：
```go
["admin", "2", "/user/list", "POST", "allow"]
["admin", "2", "/role/create", "POST", "allow"]
["admin", "2", "/department/list", "POST", "allow"]
```

### 3. 实现逻辑

```go
func (p *CoreTenantPlugin) initAdminAPIPermissions(ctx context.Context, adminRole *ent.Role, tenantID uint64) error {
    // 1. 使用系统上下文查询所有API
    systemCtx := hooks.NewSystemContext(ctx)
    apis, err := p.svcCtx.DB.API.Query().All(systemCtx)

    // 2. 构建权限策略
    var adminPolicies [][]string
    tenantIDStr := fmt.Sprintf("%d", tenantID)
    for _, api := range apis {
        adminPolicies = append(adminPolicies, []string{
            adminRole.Code,  // sub
            tenantIDStr,     // domain
            api.Path,        // obj
            api.Method,      // act
            "allow",         // eft
        })
    }

    // 3. 清除旧策略
    csb := p.svcCtx.Config.CasbinConf.MustNewCasbinWithOriginalRedisWatcher(...)
    csb.RemoveFilteredPolicy(0, adminRole.Code, tenantIDStr)

    // 4. 添加新策略
    csb.AddPolicies(adminPolicies)
}
```

## 📊 初始化流程对比

### 修改前：

```
1. 初始化字典数据
2. 初始化系统配置
3. 初始化租户菜单
4. 创建默认部门和职位
5. 创建管理员角色和用户
6. 完成初始化 ❌ 管理员无API权限
```

### 修改后：

```
1. 初始化字典数据
2. 初始化系统配置
3. 初始化租户菜单
4. 创建默认部门和职位
5. 创建管理员角色和用户
6. ✅ 初始化API权限（新增）
7. 完成初始化 ✅ 管理员拥有全部API权限
```

## 🔍 日志输出

成功初始化时会输出以下日志：

```json
{
  "level": "info",
  "msg": "Initializing API permissions for admin role",
  "tenant_id": 2,
  "role_id": 5,
  "role_code": "admin",
  "api_count": 165
}

{
  "level": "info",
  "msg": "Successfully initialized API permissions for admin role",
  "tenant_id": 2,
  "role_id": 5,
  "role_code": "admin",
  "policy_count": 165
}
```

## 🎨 关键设计点

### 1. 使用SystemContext查询API

**原因**：API表是系统级数据，不隔离租户。必须使用SystemContext绕过租户隔离。

```go
systemCtx := hooks.NewSystemContext(ctx)
apis, err := p.svcCtx.DB.API.Query().All(systemCtx)
```

### 2. 使用带Redis Watcher的Casbin Enforcer

**原因**：确保策略变更能够通知所有服务实例，实现多实例间的策略同步。

```go
csb := p.svcCtx.Config.CasbinConf.MustNewCasbinWithOriginalRedisWatcher(
    p.svcCtx.Config.DatabaseConf.Type,
    p.svcCtx.Config.DatabaseConf.GetDSN(),
    p.svcCtx.Config.RedisConf)
```

### 3. 清除旧策略

**原因**：避免重复初始化时产生重复策略，确保策略的唯一性。

```go
oldPolicies, err := csb.GetFilteredPolicy(0, adminRole.Code, tenantIDStr)
if len(oldPolicies) > 0 {
    csb.RemoveFilteredPolicy(0, adminRole.Code, tenantIDStr)
}
```

### 4. 在事务中执行

**原因**：确保角色创建、用户创建、权限初始化都在同一事务中，要么全部成功，要么全部失败。

```go
tx, err := p.svcCtx.DB.Tx(tenantCtx)
// ... 所有操作 ...
tx.Commit()
```

## 📝 参考实现

本实现参考了 `init_database_logic.go` 中的 `insertCasbinPoliciesData` 方法：

| 功能 | init_database | 租户初始化 |
|------|--------------|-----------|
| 查询API | ✅ 系统上下文 | ✅ 系统上下文 |
| 角色代码 | superadmin | admin（动态） |
| 租户ID | 1（固定） | 动态传入 |
| 策略格式 | 5字段RBAC | 5字段RBAC |
| Redis Watcher | ✅ 使用 | ✅ 使用 |

## 🧪 测试验证

### 1. 创建新租户

```bash
curl -X POST http://localhost:9100/tenant/init \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 2,
    "admin_username": "admin",
    "admin_password": "123456",
    "admin_email": "admin@tenant2.com"
  }'
```

### 2. 验证权限数据

```sql
-- 查询Casbin策略表
SELECT * FROM sys_casbin_rules
WHERE v0 = 'admin' AND v1 = '2';

-- 应该看到165条记录（对应165个API）
```

### 3. 登录测试

```bash
# 使用新租户管理员登录
curl -X POST http://localhost:3100/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "123456",
    "tenant_id": 2
  }'

# 登录成功后，应该能够访问所有API接口
```

## ⚠️ 注意事项

1. **API表必须已初始化**：租户初始化前，系统必须已通过 `init_database` 初始化了API数据。

2. **角色代码一致性**：管理员角色的 `code` 字段必须是 "admin"，与Casbin策略中的 `sub` 对应。

3. **租户ID格式**：Casbin策略中的租户ID必须是字符串格式。

4. **并发安全**：使用Redis Watcher确保多实例环境下的策略同步。

5. **事务回滚**：任何阶段失败都会回滚整个事务，不会产生不一致状态。

## 🚀 部署更新

```bash
# 1. 重新编译RPC服务
cd /opt/code/newbee/core/rpc
go build -v -o /tmp/core.rpc .

# 2. 停止旧服务
lsof -ti:9100 | xargs kill -9

# 3. 启动新服务
/tmp/core.rpc > /tmp/core_rpc.log 2>&1 &

# 4. 验证启动成功
tail -f /tmp/core_rpc.log
```

## 📌 版本信息

- **功能版本**：v2.0.0
- **修改日期**：2025-10-09
- **修改文件**：
  - `core/rpc/internal/plugins/core_plugin.go`
  - `core/rpc/internal/plugins/core_plugin_methods.go`

---

**作者**：Claude Code
**最后更新**：2025-10-09 03:40
