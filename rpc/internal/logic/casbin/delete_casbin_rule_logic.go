package casbin

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCasbinRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCasbinRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCasbinRuleLogic {
	return &DeleteCasbinRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCasbinRuleLogic) DeleteCasbinRule(in *core.IDsReq) (*core.BaseResp, error) {
	// 验证IDs
	if len(in.Ids) == 0 {
		return nil, fmt.Errorf("ids cannot be empty")
	}

	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)
	
	// 查找要删除的规则 - 必须包含租户ID过滤
	rules, err := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.IDIn(in.Ids...),
		).
		All(l.ctx)
	if err != nil {
		l.Logger.Errorf("Query casbin rules failed: %v", err)
		return nil, fmt.Errorf("query casbin rules failed: %v", err)
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("no casbin rules found with the provided IDs")
	}

	// 🔥 先从 Casbin 引擎移除规则
	err = l.syncRemoveFromCasbinEngine(rules)
	if err != nil {
		// 记录警告但继续执行数据库删除
		l.Logger.Errorf("Remove rules from Casbin engine failed: %v", err)
	}

	// 记录将要删除的规则信息
	var deletedRules []string
	for _, rule := range rules {
		deletedRules = append(deletedRules, fmt.Sprintf("ID:%d,ptype:%s,service:%s", 
			rule.ID, rule.Ptype, rule.ServiceName))
	}

	// 执行批量删除 - 必须包含租户ID过滤
	deletedCount, err := l.svcCtx.DB.CasbinRule.Delete().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离 - 关键安全控制
			casbinrule.IDIn(in.Ids...),
		).
		Exec(l.ctx)
	if err != nil {
		l.Logger.Errorf("Delete casbin rules failed: %v", err)
		return nil, fmt.Errorf("delete casbin rules failed: %v", err)
	}

	l.Logger.Infof("Deleted %d casbin rules successfully. Rules: %v", 
		deletedCount, deletedRules)

	return &core.BaseResp{
		Msg: fmt.Sprintf("成功删除 %d 条权限规则", deletedCount),
	}, nil
}

// syncRemoveFromCasbinEngine 从 Casbin 引擎批量移除规则
func (l *DeleteCasbinRuleLogic) syncRemoveFromCasbinEngine(rules []*ent.CasbinRule) error {
	// 检查 EnforcerManager 是否可用
	if l.svcCtx.EnforcerManager == nil {
		return fmt.Errorf("EnforcerManager not initialized")
	}

	// 逐个移除规则
	var lastErr error
	successCount := 0
	for _, rule := range rules {
		err := l.removeRuleFromCasbinEngine(rule)
		if err != nil {
			l.Logger.Errorf("Remove rule from Casbin engine failed: %v, rule ID: %d", err, rule.ID)
			lastErr = err
		} else {
			successCount++
		}
	}

	l.Logger.Infof("Removed %d/%d rules from Casbin engine", successCount, len(rules))
	return lastErr
}

// removeRuleFromCasbinEngine 从 Casbin 引擎移除单个规则
// 🔥 适配 RBAC with Domains 模型
func (l *DeleteCasbinRuleLogic) removeRuleFromCasbinEngine(rule *ent.CasbinRule) error {
	// 🔥 根据 Ptype 类型从 Casbin 引擎移除
	switch rule.Ptype {
	case "p":
		// 🔥 新格式：v0=sub, v1=domain, v2=obj, v3=act
		// EnforcerManager.RemovePolicy 期望：sub, obj, act（内部自动加domain）
		if rule.V0 != "" && rule.V2 != "" && rule.V3 != "" {
			_, err := l.svcCtx.EnforcerManager.RemovePolicy(l.ctx, rule.TenantID, rule.V0, rule.V2, rule.V3)
			return err
		}
	case "g":
		// 🔥 新格式：v0=user, v1=role, v2=domain
		// EnforcerManager.RemoveGroupingPolicy 期望：user, role（内部自动加domain）
		if rule.V0 != "" && rule.V1 != "" {
			_, err := l.svcCtx.EnforcerManager.RemoveGroupingPolicy(l.ctx, rule.TenantID, rule.V0, rule.V1)
			return err
		}
	case "g2", "g3", "g4", "g5":
		// 🔥 扩展角色继承：v0=user, v1=role, v2=domain
		if rule.V0 != "" && rule.V1 != "" {
			_, err := l.svcCtx.EnforcerManager.RemoveNamedGroupingPolicy(l.ctx, rule.TenantID, rule.Ptype, rule.V0, rule.V1)
			return err
		}
	default:
		l.Logger.Infof("Unknown ptype: %s, skipping Casbin engine sync", rule.Ptype)
	}

	return nil
}

// buildCasbinParams 构建 Casbin 参数
// 🔥 注意：此方法已废弃，保留仅为兼容性
// 新的实现直接在 removeRuleFromCasbinEngine 中处理
func (l *DeleteCasbinRuleLogic) buildCasbinParams(rule *ent.CasbinRule) []string {
	params := []string{}
	if rule.V0 != "" {
		params = append(params, rule.V0)
	}
	if rule.V1 != "" {
		params = append(params, rule.V1)
	}
	if rule.V2 != "" {
		params = append(params, rule.V2)
	}
	if rule.V3 != "" {
		params = append(params, rule.V3)
	}
	if rule.V4 != "" {
		params = append(params, rule.V4)
	}
	if rule.V5 != "" {
		params = append(params, rule.V5)
	}
	return params
}
