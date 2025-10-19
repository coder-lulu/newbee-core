package casbin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	casbinMgr "github.com/coder-lulu/newbee-core/rpc/internal/casbin"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 权限验证
func (l *CheckPermissionLogic) CheckPermission(in *core.PermissionCheckReq) (*core.PermissionCheckResp, error) {
	// 验证请求参数
	if in.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}
	if in.Subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	if in.Object == "" {
		return nil, fmt.Errorf("object is required")
	}
	if in.Action == "" {
		return nil, fmt.Errorf("action is required")
	}

	startTime := time.Now()
	var fromCache bool

	// 使用 Casbin 引擎进行权限检查（带缓存）
	result, err := l.svcCtx.EnforcerManager.CheckPermissionWithRoles(l.ctx, in.Subject, in.Object, in.Action, in.ServiceName)
	if err != nil {
		// 如果Casbin检查失败，降级到数据库查询
		l.Logger.Errorf("Casbin permission check failed, fallback to DB: %v", err)
		allowed, appliedRules, reason, dbErr := l.checkPermissionInDB(in)
		if dbErr != nil {
			l.Logger.Errorf("Permission check failed: %v", dbErr)
			return nil, fmt.Errorf("permission check failed: %v", dbErr)
		}
		result = &casbinMgr.PermissionResult{
			Allowed:      allowed,
			Reason:       reason,
			AppliedRules: appliedRules,
			FromCache:    false,
		}
	}

	fromCache = result.FromCache

	// 记录审计日志
	if in.AuditLog != nil && *in.AuditLog {
		l.logPermissionCheck(in, result.Allowed, result.Reason)
	}

	duration := time.Since(startTime).Milliseconds()

	return &core.PermissionCheckResp{
		Allowed:         result.Allowed,
		Reason:          result.Reason,
		AppliedRules:    result.AppliedRules,
		DataFilters:     make(map[string]string), // TODO: 实现数据过滤
		FieldMasks:      []string{},               // TODO: 实现字段掩码
		CheckDurationMs: duration,
		FromCache:       fromCache,
	}, nil
}

// checkPermissionInDB 使用数据库直接查询进行权限检查
func (l *CheckPermissionLogic) checkPermissionInDB(in *core.PermissionCheckReq) (bool, []string, string, error) {
	now := time.Now()
	
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)
	
	// 查询匹配的规则 - 必须包含租户ID过滤
	rules, err := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID),       // 🔥 租户隔离 - 关键安全控制
			casbinrule.ServiceNameEQ(in.ServiceName),
			casbinrule.V0EQ(in.Subject),           // 主体
			casbinrule.V1EQ(in.Object),            // 资源
			casbinrule.V2EQ(in.Action),            // 操作
			casbinrule.StatusEQ(1),                // 启用状态
			casbinrule.Or(
				casbinrule.EffectiveFromLTE(now),    // 生效时间
				casbinrule.EffectiveFromIsNil(),
			),
			casbinrule.Or(
				casbinrule.EffectiveToGTE(now),      // 失效时间
				casbinrule.EffectiveToIsNil(),
			),
		).
		All(l.ctx)

	if err != nil {
		return false, nil, "database query failed", err
	}

	if len(rules) == 0 {
		return false, nil, "no matching permission rules found", nil
	}

	// 检查是否有待审批的规则
	var appliedRuleIDs []string
	var allowRules, denyRules int

	for _, rule := range rules {
		// 检查审批状态
		if rule.RequireApproval && rule.ApprovalStatus != "approved" {
			continue // 跳过未审批的规则
		}

		appliedRuleIDs = append(appliedRuleIDs, fmt.Sprintf("%d", rule.ID))

		// 判断效果
		switch strings.ToLower(rule.V3) {
		case "allow", "":
			allowRules++
		case "deny":
			denyRules++
		}
	}

	// 决策逻辑：如果有拒绝规则，则拒绝访问
	if denyRules > 0 {
		return false, appliedRuleIDs, fmt.Sprintf("access denied by %d rule(s)", denyRules), nil
	}

	if allowRules > 0 {
		return true, appliedRuleIDs, fmt.Sprintf("access granted by %d rule(s)", allowRules), nil
	}

	return false, appliedRuleIDs, "no effective rules found", nil
}

// checkPermissionWithCasbin 使用Casbin引擎进行权限检查（带缓存）
func (l *CheckPermissionLogic) checkPermissionWithCasbin(in *core.PermissionCheckReq) (bool, []string, string, error) {
	// 使用增强权限检查（包含角色和缓存）
	result, err := l.svcCtx.EnforcerManager.CheckPermissionWithRoles(l.ctx, in.Subject, in.Object, in.Action, in.ServiceName)
	if err != nil {
		return false, nil, fmt.Sprintf("casbin check error: %v", err), err
	}

	return result.Allowed, result.AppliedRules, result.Reason, nil
}

// logPermissionCheck 记录权限检查审计日志
func (l *CheckPermissionLogic) logPermissionCheck(req *core.PermissionCheckReq, allowed bool, reason string) {
	result := "DENIED"
	if allowed {
		result = "ALLOWED"
	}

	l.Logger.Infof("Permission check audit: service=%s, subject=%s, object=%s, action=%s, result=%s, reason=%s",
		req.ServiceName, req.Subject, req.Object, req.Action, result, reason)
}
