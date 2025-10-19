# Context Metadata 提取工具

## 概述

`contextx` 包提供了统一的工具函数，用于从 gRPC context 或 metadata 中提取认证和授权信息。

## 为什么需要这个工具？

在 gRPC 调用链中，认证信息（如 tenant_id, user_id, dept_id, role_codes 等）通过以下两种方式传递：

1. **Context Values** - 在同一服务内传递
2. **gRPC Metadata** - 在跨服务调用时传递

由于 API 服务调用 RPC 服务时使用 gRPC metadata，RPC 服务的 logic 层需要从 metadata 中提取这些信息。

## 使用方法

### 1. 提取单个字段

```go
import "github.com/coder-lulu/newbee-core/rpc/internal/utils/contextx"

// 提取租户ID
tenantID := contextx.ExtractTenantID(ctx)

// 提取用户ID
userID := contextx.ExtractUserID(ctx)

// 提取部门ID
deptID := contextx.ExtractDeptID(ctx)

// 提取角色代码
roleCodes := contextx.ExtractRoleCodes(ctx)

// 提取数据权限范围
dataScope := contextx.ExtractDataScope(ctx)
```

### 2. 提取所有认证信息

```go
authInfo := contextx.ExtractAllAuthInfo(ctx)

tenantID := authInfo["tenant_id"]
userID := authInfo["user_id"]
deptID := authInfo["dept_id"]
roleCodes := authInfo["role_codes"]
dataScope := authInfo["data_scope"]
```

### 3. 完整示例

```go
package user

import (
    "context"
    "strings"

    "github.com/coder-lulu/newbee-core/rpc/internal/svc"
    "github.com/coder-lulu/newbee-core/rpc/internal/utils/contextx"
    "github.com/coder-lulu/newbee-core/rpc/types/core"

    "github.com/zeromicro/go-zero/core/errorx"
    "github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
    logx.Logger
}

func (l *GetUserListLogic) GetUserList(in *core.UserListReq) (*core.UserListResp, error) {
    // 提取所有认证信息
    authInfo := contextx.ExtractAllAuthInfo(l.ctx)

    // 记录日志
    logx.Infow("Processing user list request",
        logx.Field("tenant_id", authInfo["tenant_id"]),
        logx.Field("user_id", authInfo["user_id"]),
        logx.Field("dept_id", authInfo["dept_id"]),
        logx.Field("role_codes", authInfo["role_codes"]))

    // 权限检查示例
    roleCodes := authInfo["role_codes"]
    if roleCodes == "" {
        return nil, errorx.NewCodeInvalidArgumentError("缺少角色信息")
    }

    roles := strings.Split(roleCodes, ",")
    hasAdminRole := false
    for _, role := range roles {
        if role == "admin" || role == "superadmin" {
            hasAdminRole = true
            break
        }
    }

    if !hasAdminRole {
        return nil, errorx.NewCodeInvalidArgumentError("权限不足")
    }

    // 业务逻辑...
    return &core.UserListResp{}, nil
}
```

## 工作原理

`ExtractFromContext` 函数按以下优先级提取值：

1. 首先尝试从 `context.Value(key)` 中获取（同一服务内传递）
2. 如果失败，尝试从 `metadata.FromIncomingContext()` 中提取（跨服务传递）
3. 如果都失败，返回空字符串

## 注意事项

1. **字段为空不一定是错误** - 有些字段（如 dept_id, data_scope）可能在某些场景下为空
2. **必需字段验证** - 对于业务必需的字段（如 tenant_id），应该在使用前进行验证
3. **性能考虑** - `ExtractAllAuthInfo` 会提取所有字段，如果只需要一两个字段，建议使用单独的提取函数

## 相关文件

- **Auth Plugin**: `/opt/code/newbee/common/middleware/auth/plugin.go` - 设置 metadata
- **gRPC Interceptor**: `/opt/code/newbee/common/orm/ent/hooks/grpc_interceptor.go` - 传递 metadata
- **Context Keys**: `/opt/code/newbee/common/middleware/keys/context_keys.go` - 定义所有 key
