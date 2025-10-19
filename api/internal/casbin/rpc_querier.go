package casbin

import (
	"context"
	"fmt"
	"strconv"

	commontypes "github.com/coder-lulu/newbee-common/v2/casbin/types"
	"github.com/coder-lulu/newbee-common/v2/middleware/dataperm"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	coreclient "github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
)

// RpcCasbinRuleQuerier API服务端的Casbin规则查询器
// 通过RPC调用从Core RPC服务查询规则
type RpcCasbinRuleQuerier struct {
	coreRpc coreclient.Core
}

// NewRpcCasbinRuleQuerier 创建RPC查询器
func NewRpcCasbinRuleQuerier(coreRpc coreclient.Core) *RpcCasbinRuleQuerier {
	return &RpcCasbinRuleQuerier{
		coreRpc: coreRpc,
	}
}

// QueryCasbinRules 实现CasbinRuleQuerier接口
// 通过RPC从Core服务查询Casbin规则
//
// 🔥 关于tenantID参数的处理说明：
// - 接口定义包含tenantID参数以保持通用性（直接数据库查询时需要）
// - 但在RPC调用场景中，租户过滤由以下机制处理：
//   1. 如果ctx包含tenantID（普通请求），RPC端的Hook会自动过滤
//   2. 如果ctx是SystemContext（初始化时），会获取所有租户的规则
// - 当前API服务使用SystemContext初始化，加载所有租户规则
// - 运行时通过RBAC with Domains模型的domain参数实现租户隔离
func (q *RpcCasbinRuleQuerier) QueryCasbinRules(ctx context.Context, tenantID uint64) ([]commontypes.CasbinRuleEntity, error) {
	// 调用RPC获取Casbin规则列表
	// 🔥 使用较大的pageSize确保获取所有规则（通常单租户不会超过10000条）
	// 注意：GetCasbinRuleList接口没有tenantID字段，租户过滤由ctx和Hook控制
	resp, err := q.coreRpc.GetCasbinRuleList(ctx, &core.CasbinRuleListReq{
		Page:     1,
		PageSize: 10000, // 足够大的页大小
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query casbin rules via RPC: %w", err)
	}

	// 转换RPC响应为通用接口
	result := make([]commontypes.CasbinRuleEntity, 0, len(resp.Data))
	for _, rule := range resp.Data {
		result = append(result, &RpcCasbinRuleWrapper{rule: rule})
	}

	return result, nil
}

// RpcCasbinRuleWrapper 包装RPC响应数据以实现CasbinRuleEntity接口
type RpcCasbinRuleWrapper struct {
	rule *core.CasbinRuleInfo
}

// 实现CasbinRuleEntity接口的所有方法

func (w *RpcCasbinRuleWrapper) GetID() uint64 {
	if w.rule.Id != nil {
		return *w.rule.Id
	}
	return 0
}

func (w *RpcCasbinRuleWrapper) GetPtype() string {
	return w.rule.Ptype
}

func (w *RpcCasbinRuleWrapper) GetV0() string {
	if w.rule.V0 != nil {
		return *w.rule.V0
	}
	return ""
}

func (w *RpcCasbinRuleWrapper) GetV1() string {
	if w.rule.V1 != nil {
		return *w.rule.V1
	}
	return ""
}

func (w *RpcCasbinRuleWrapper) GetV2() string {
	if w.rule.V2 != nil {
		return *w.rule.V2
	}
	return ""
}

func (w *RpcCasbinRuleWrapper) GetV3() string {
	if w.rule.V3 != nil {
		return *w.rule.V3
	}
	return ""
}

func (w *RpcCasbinRuleWrapper) GetV4() string {
	if w.rule.V4 != nil {
		return *w.rule.V4
	}
	return ""
}

func (w *RpcCasbinRuleWrapper) GetV5() string {
	if w.rule.V5 != nil {
		return *w.rule.V5
	}
	return ""
}

func (w *RpcCasbinRuleWrapper) GetTenantID() uint64 {
	if w.rule.TenantId != nil {
		return *w.rule.TenantId
	}
	return 0
}

func (w *RpcCasbinRuleWrapper) GetStatus() uint8 {
	if w.rule.Status != nil {
		return uint8(*w.rule.Status)
	}
	return 0
}

func (w *RpcCasbinRuleWrapper) GetRequireApproval() bool {
	if w.rule.RequireApproval != nil {
		return *w.rule.RequireApproval
	}
	return false
}

func (w *RpcCasbinRuleWrapper) GetApprovalStatus() string {
	if w.rule.ApprovalStatus != nil {
		return *w.rule.ApprovalStatus
	}
	return "pending"
}

func (w *RpcCasbinRuleWrapper) HasEffectiveFrom() bool {
	// 检查effectiveFrom是否有值（非0）
	return w.rule.EffectiveFrom != nil && *w.rule.EffectiveFrom > 0
}

func (w *RpcCasbinRuleWrapper) HasEffectiveTo() bool {
	// 检查effectiveTo是否有值（非0）
	return w.rule.EffectiveTo != nil && *w.rule.EffectiveTo > 0
}

// GetEffectiveFrom 获取生效时间（扩展方法，非接口要求）
func (w *RpcCasbinRuleWrapper) GetEffectiveFrom() int64 {
	if w.rule.EffectiveFrom != nil {
		return *w.rule.EffectiveFrom
	}
	return 0
}

// GetEffectiveTo 获取失效时间（扩展方法，非接口要求）
func (w *RpcCasbinRuleWrapper) GetEffectiveTo() int64 {
	if w.rule.EffectiveTo != nil {
		return *w.rule.EffectiveTo
	}
	return 0
}

// String 实现Stringer接口，方便调试
func (w *RpcCasbinRuleWrapper) String() string {
	return fmt.Sprintf("CasbinRule{ID:%d, Ptype:%s, V0:%s, V1:%s, V2:%s, V3:%s, Status:%d, TenantID:%d}",
		w.GetID(), w.GetPtype(), w.GetV0(), w.GetV1(), w.GetV2(), w.GetV3(),
		w.GetStatus(), w.GetTenantID())
}

// GetDomain 获取domain（租户ID的字符串形式）
func (w *RpcCasbinRuleWrapper) GetDomain() string {
	return strconv.FormatUint(w.GetTenantID(), 10)
}

// ============================================
// 🔥 Phase 2: 实现CasbinProvider接口
// ============================================

// CheckPermissionWithRoles 实现dataperm.CasbinProvider接口 - 检查权限（包含角色支持）
// 通过RPC调用Core服务的CheckPermission方法
func (q *RpcCasbinRuleQuerier) CheckPermissionWithRoles(
	ctx context.Context,
	subject, object, action, serviceName string,
) (*dataperm.PermissionResult, error) {
	// 构建RPC请求
	req := &core.PermissionCheckReq{
		ServiceName: serviceName,
		Subject:     subject,
		Object:      object,
		Action:      action,
		Context:     nil,
		EnableCache: pointy.GetPointer(true),  // 启用缓存
		AuditLog:    pointy.GetPointer(false), // 不记录审计日志（避免循环）
	}

	// 调用Core RPC服务
	resp, err := q.coreRpc.CheckPermission(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission via RPC: %w", err)
	}

	// 转换为dataperm.PermissionResult
	result := &dataperm.PermissionResult{
		Allowed:      resp.Allowed,
		Reason:       resp.Reason,
		AppliedRules: resp.AppliedRules,
		FromCache:    resp.FromCache,
	}

	return result, nil
}

// GetUserRolesWithCache 实现dataperm.CasbinProvider接口 - 获取用户角色（带缓存）
// 通过RPC调用Core服务的GetUserById方法获取用户角色
func (q *RpcCasbinRuleQuerier) GetUserRolesWithCache(ctx context.Context, user string) ([]string, error) {
	// 构建RPC请求
	req := &core.UUIDReq{
		Id: user,
	}

	// 调用Core RPC服务
	userInfo, err := q.coreRpc.GetUserById(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info via RPC: %w", err)
	}

	// 返回角色代码列表
	if len(userInfo.RoleCodes) == 0 {
		return []string{}, nil
	}

	return userInfo.RoleCodes, nil
}
