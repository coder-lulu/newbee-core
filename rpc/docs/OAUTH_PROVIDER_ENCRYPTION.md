# OAuth Provider 加密功能使用指南

## 1. 功能概述

OAuth Provider 加密功能为 `client_secret` 提供企业级加密保护，使用 **AES-256-GCM** 算法确保敏感数据在数据库中以加密形式存储。

### 1.1 核心特性

- ✅ **AES-256-GCM 加密** - 军事级加密算法
- ✅ **自动加解密** - 业务逻辑透明处理
- ✅ **密钥轮换支持** - 定期密钥更新机制
- ✅ **向后兼容** - 支持旧数据平滑迁移
- ✅ **数据迁移工具** - 一键加密现有数据
- ✅ **加密状态验证** - 实时监控加密覆盖率

## 2. 架构设计

### 2.1 加密流程

```
创建/更新 Provider
    ↓
client_secret (明文)
    ↓
EncryptionService.EncryptProviderSecret()
    ↓
encrypted_secret (密文) + encryption_key_id
    ↓
存储到数据库 (清除明文)
```

### 2.2 解密流程

```
OAuth 登录/回调
    ↓
从数据库获取 encrypted_secret + encryption_key_id
    ↓
EncryptionService.DecryptProviderSecret()
    ↓
client_secret (临时明文)
    ↓
用于 OAuth API 调用
```

### 2.3 数据库字段

| 字段名 | 类型 | 说明 |
|-------|------|------|
| `client_secret` | string | (已废弃) 明文密钥，加密后清空 |
| `encrypted_secret` | string | Base64编码的加密密钥 |
| `encryption_key_id` | string | 加密密钥版本ID (如: "prod-2024") |

## 3. 配置步骤

### 3.1 配置文件设置

编辑 `/opt/code/newbee/core/rpc/etc/core.yaml`:

```yaml
# 🔐 OAuth Provider加密密钥配置
EncryptionKey: "your-strong-32-byte-encryption-key-here"  # 必须32字节或更长
```

**⚠️ 安全建议**:
- 生产环境必须使用强随机密钥 (至少32字节)
- 不要在代码中硬编码密钥
- 建议使用环境变量或密钥管理服务 (KMS)
- 定期轮换密钥 (建议3-6个月)

### 3.2 密钥生成示例

```bash
# 方法1: OpenSSL 生成32字节随机密钥
openssl rand -base64 32

# 方法2: Go 生成
go run -<<'EOF'
package main
import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
)
func main() {
    key := make([]byte, 32)
    rand.Read(key)
    fmt.Println(base64.StdEncoding.EncodeToString(key))
}
EOF
```

## 4. 使用场景

### 4.1 创建新 Provider (自动加密)

```go
// 前端提交
req := &core.OauthProviderInfo{
    Name:         pointy.String("google"),
    ClientId:     pointy.String("xxx.apps.googleusercontent.com"),
    ClientSecret: pointy.String("GOCSPX-xxxxxxxxxx"),  // 明文
    // ... 其他字段
}

// 后端自动加密
l := NewCreateOauthProviderLogic(ctx, svcCtx)
resp, err := l.CreateOauthProvider(req)

// ✅ 数据库存储:
// client_secret = ""
// encrypted_secret = "base64_encrypted_data"
// encryption_key_id = "prod-2024"
```

### 4.2 更新 Provider (自动加密)

```go
// 更新密钥时自动重新加密
req := &core.OauthProviderInfo{
    Id:           pointy.Uint64(1),
    ClientSecret: pointy.String("new-secret-value"),  // 新密钥
}

l := NewUpdateOauthProviderLogic(ctx, svcCtx)
resp, err := l.UpdateOauthProvider(req)
```

### 4.3 查询 Provider (自动解密)

**管理接口 - 返回解密后的密钥**:
```go
req := &core.IDReq{Id: 1}
l := NewGetOauthProviderByIdLogic(ctx, svcCtx)
resp, err := l.GetOauthProviderById(req)

// ✅ resp.ClientSecret 返回解密后的明文 (仅用于管理界面)
```

**列表接口 - 返回掩码**:
```go
req := &core.OauthProviderListReq{Page: 1, PageSize: 10}
l := NewGetOauthProviderListLogic(ctx, svcCtx)
resp, err := l.GetOauthProviderList(req)

// ✅ resp.Data[i].ClientSecret 返回 "******" (安全掩码)
```

### 4.4 OAuth 登录/回调 (自动解密)

```go
// OAuth 登录流程
req := &core.OauthLoginReq{Provider: "google"}
l := NewOauthLoginLogic(ctx, svcCtx)
resp, err := l.OauthLogin(req)

// ✅ 内部自动解密 client_secret 用于生成 OAuth URL

// OAuth 回调流程
req := &core.CallbackReq{Code: "xxx", State: "xxx"}
l := NewOauthCallbackLogic(ctx, svcCtx)
resp, err := l.OauthCallback(req)

// ✅ 内部自动解密 client_secret 用于 token exchange
```

## 5. 数据迁移

### 5.1 迁移现有数据

```go
import "github.com/coder-lulu/newbee-core/rpc/internal/logic/oauthprovider"

// 一次性迁移所有明文数据
err := oauthprovider.MigrateEncryption(ctx, svcCtx)
if err != nil {
    log.Fatalf("Migration failed: %v", err)
}

// ✅ 自动:
// 1. 查询所有 client_secret 不为空但 encrypted_secret 为空的记录
// 2. 加密 client_secret
// 3. 存储 encrypted_secret + encryption_key_id
// 4. 清除 client_secret 明文
```

### 5.2 验证加密状态

```go
stats, err := oauthprovider.ValidateEncryption(ctx, svcCtx)
if err != nil {
    log.Fatalf("Validation failed: %v", err)
}

// 返回统计信息:
// {
//     "total_providers": 10,
//     "encrypted_providers": 8,
//     "unencrypted_providers": 2,
//     "encryption_rate": 80.0,
//     "all_encrypted": false
// }
```

## 6. 密钥轮换

### 6.1 单个 Provider 轮换

```go
// 重新加密单个 Provider (使用新密钥)
err := oauthprovider.ReEncryptProvider(ctx, svcCtx, providerID)
if err != nil {
    log.Fatalf("Re-encryption failed: %v", err)
}

// ✅ 自动:
// 1. 解密旧密钥
// 2. 使用当前活跃密钥重新加密
// 3. 更新数据库
```

### 6.2 批量轮换所有密钥

```go
// 批量轮换所有 Provider (定期执行)
err := oauthprovider.RotateAllKeys(ctx, svcCtx)
if err != nil {
    log.Fatalf("Batch rotation failed: %v", err)
}

// ✅ 适用场景:
// - 定期密钥轮换 (每3-6个月)
// - 密钥泄露应急响应
// - 升级加密算法
```

### 6.3 密钥轮换最佳实践

**步骤1: 准备新密钥**
```yaml
# core.yaml
EncryptionKey: "new-strong-32-byte-key-2025"  # 新密钥
```

**步骤2: 更新配置并重启服务**
```bash
# 重启 RPC 服务以加载新密钥
systemctl restart core-rpc
```

**步骤3: 执行批量轮换**
```go
// 在管理接口或定时任务中执行
err := oauthprovider.RotateAllKeys(ctx, svcCtx)
```

**步骤4: 验证**
```go
stats, _ := oauthprovider.ValidateEncryption(ctx, svcCtx)
// 确认 encryption_rate = 100%
```

## 7. 安全最佳实践

### 7.1 密钥管理

- ✅ 使用强随机密钥 (32字节以上)
- ✅ 密钥存储在安全配置中心 (不提交到代码仓库)
- ✅ 定期轮换密钥 (3-6个月)
- ✅ 保留旧密钥直到所有数据轮换完成
- ❌ 不要在日志中打印密钥
- ❌ 不要在错误信息中暴露密钥

### 7.2 访问控制

- ✅ 管理接口 (GetById) 仅限管理员访问
- ✅ 列表接口 (GetList) 返回掩码值 "******"
- ✅ OAuth 流程内部自动解密，不暴露给前端
- ❌ 不要在 API 响应中返回明文密钥

### 7.3 审计日志

所有加密/解密操作已记录详细日志:

```go
// 成功日志
logger.Infow("Successfully encrypted provider secret", 
    logx.Field("provider_id", p.ID),
    logx.Field("key_id", keyID))

// 失败日志
logger.Errorw("Failed to decrypt client secret",
    logx.Field("error", err),
    logx.Field("provider", p.Name))
```

## 8. 故障排查

### 8.1 常见问题

**问题1: 解密失败 "decryption failed"**

原因: 密钥配置错误或已更改

解决:
1. 检查 `core.yaml` 中的 `EncryptionKey` 是否正确
2. 确认密钥长度至少32字节
3. 验证 `encryption_key_id` 对应的密钥是否存在

**问题2: OAuth 登录失败 "invalid client"**

原因: client_secret 解密后不正确

解决:
1. 使用 `GetOauthProviderById` 验证解密后的密钥
2. 对比 OAuth Provider 控制台中的正确密钥
3. 必要时重新更新密钥

**问题3: 迁移失败 "migration completed with failures"**

原因: 部分 Provider 加密失败

解决:
1. 查看详细日志定位失败的 Provider
2. 检查失败 Provider 的 `client_secret` 是否有效
3. 手动修复后重新运行迁移

### 8.2 日志分析

**启用详细日志**:
```yaml
# core.yaml
Log:
  Level: debug  # 开启调试日志
```

**关键日志标识**:
- `Successfully encrypted provider secret` - 加密成功
- `Failed to decrypt client secret` - 解密失败
- `Starting encryption migration` - 迁移开始
- `Batch key rotation completed` - 轮换完成

## 9. 性能考虑

### 9.1 性能指标

- 加密操作: ~0.1ms (AES-256-GCM)
- 解密操作: ~0.1ms
- 批量迁移: ~10 providers/s

### 9.2 优化建议

- ✅ 列表接口返回掩码，避免批量解密
- ✅ OAuth 流程仅解密需要的 Provider
- ✅ 使用缓存存储 OAuth Config (已实现)
- ❌ 不要在循环中重复解密相同密钥

## 10. 监控与告警

### 10.1 关键指标

- `encrypted_providers` - 已加密 Provider 数量
- `unencrypted_providers` - 未加密 Provider 数量 (应为0)
- `encryption_rate` - 加密覆盖率 (应为100%)
- 解密失败次数 - 应为0

### 10.2 告警规则

```yaml
# Prometheus 告警示例
- alert: UnencryptedProviders
  expr: oauth_unencrypted_providers > 0
  for: 5m
  annotations:
    summary: "发现未加密的 OAuth Provider"
    
- alert: DecryptionFailures
  expr: rate(oauth_decryption_failures[5m]) > 0
  annotations:
    summary: "OAuth 密钥解密失败"
```

## 11. API 参考

### 11.1 加密服务接口

```go
type ProviderEncryptionService interface {
    // 加密 client_secret
    EncryptProviderSecret(plaintext string) (encrypted string, keyID string, err error)
    
    // 解密 client_secret
    DecryptProviderSecret(encrypted string, keyID string) (plaintext string, err error)
}
```

### 11.2 数据迁移接口

```go
// 迁移所有明文数据
func MigrateEncryption(ctx context.Context, svcCtx *svc.ServiceContext) error

// 验证加密状态
func ValidateEncryption(ctx context.Context, svcCtx *svc.ServiceContext) (map[string]interface{}, error)

// 单个 Provider 重新加密
func ReEncryptProvider(ctx context.Context, svcCtx *svc.ServiceContext, providerID uint64) error

// 批量密钥轮换
func RotateAllKeys(ctx context.Context, svcCtx *svc.ServiceContext) error
```

## 12. 部署检查清单

### 12.1 首次部署

- [ ] 配置强随机 `EncryptionKey` (32字节)
- [ ] 重启 RPC 服务加载配置
- [ ] 执行 `MigrateEncryption()` 迁移现有数据
- [ ] 执行 `ValidateEncryption()` 验证加密率100%
- [ ] 测试 OAuth 登录/回调流程
- [ ] 配置监控告警

### 12.2 密钥轮换

- [ ] 生成新的强随机密钥
- [ ] 更新 `core.yaml` 配置
- [ ] 重启服务加载新密钥
- [ ] 执行 `RotateAllKeys()` 批量轮换
- [ ] 验证所有 Provider 正常工作
- [ ] 备份旧密钥 (保留1-3个月)

### 12.3 应急响应 (密钥泄露)

1. **立即轮换密钥** (按上述流程)
2. **审计访问日志** (查找异常访问)
3. **通知安全团队**
4. **评估影响范围**
5. **更新所有环境的密钥**

## 13. 常见问答

**Q1: 加密对性能有影响吗?**

A: 影响极小 (<0.1ms)，仅在创建/更新/OAuth流程时执行，列表查询不解密。

**Q2: 如何回滚到明文存储?**

A: 不建议回滚。如必须回滚，需手动解密所有数据并清空 `encrypted_secret`。

**Q3: 支持多个加密密钥吗?**

A: 支持。通过 `encryption_key_id` 标识密钥版本，可实现密钥轮换。

**Q4: 数据库备份包含密钥吗?**

A: 数据库仅存储加密后的数据，不包含加密密钥。密钥在配置文件中独立管理。

**Q5: 如何审计密钥使用情况?**

A: 所有加密/解密操作都有详细日志，可通过日志系统审计。

---

**最后更新**: 2025-10-04  
**版本**: v1.0  
**维护者**: NewBee Security Team
