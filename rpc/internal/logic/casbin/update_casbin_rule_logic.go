package casbin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateCasbinRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCasbinRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCasbinRuleLogic {
	return &UpdateCasbinRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateCasbinRuleLogic) UpdateCasbinRule(in *core.CasbinRuleInfo) (*core.BaseResp, error) {
	// 验证ID
	if in.Id == nil || *in.Id == 0 {
		return nil, fmt.Errorf("id is required")
	}

	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)

	// 检查规则是否存在 - 必须包含租户ID过滤
	rule, err := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.IDEQ(*in.Id),
		).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("casbin rule not found")
		}
		return nil, fmt.Errorf("get casbin rule failed: %v", err)
	}

	// 构建更新器
	update := rule.Update()

	// 更新 Casbin 标准字段
	if in.Ptype != "" {
		update.SetPtype(in.Ptype)
	}
	if in.V0 != nil {
		update.SetV0(*in.V0)
	}
	if in.V1 != nil {
		update.SetV1(*in.V1)
	}
	if in.V2 != nil {
		update.SetV2(*in.V2)
	}
	if in.V3 != nil {
		update.SetV3(*in.V3)
	}
	if in.V4 != nil {
		update.SetV4(*in.V4)
	}
	if in.V5 != nil {
		update.SetV5(*in.V5)
	}

	// 更新业务扩展字段
	if in.ServiceName != "" {
		update.SetServiceName(in.ServiceName)
	}
	if in.RuleName != nil {
		update.SetRuleName(*in.RuleName)
	}
	if in.Description != nil {
		update.SetDescription(*in.Description)
	}
	if in.Category != nil {
		update.SetCategory(*in.Category)
	}
	if in.Version != nil {
		update.SetVersion(*in.Version)
	}

	// 更新审批流程字段
	if in.RequireApproval != nil {
		update.SetRequireApproval(*in.RequireApproval)
	}
	if in.ApprovalStatus != nil {
		update.SetApprovalStatus(casbinrule.ApprovalStatus(*in.ApprovalStatus))
	}
	if in.ApprovedBy != nil {
		update.SetApprovedBy(*in.ApprovedBy)
	}
	if in.ApprovedAt != nil {
		update.SetApprovedAt(time.Unix(*in.ApprovedAt, 0))
	}

	// 更新时间控制字段
	if in.EffectiveFrom != nil {
		update.SetEffectiveFrom(time.Unix(*in.EffectiveFrom, 0))
	}
	if in.EffectiveTo != nil {
		update.SetEffectiveTo(time.Unix(*in.EffectiveTo, 0))
	}
	if in.IsTemporary != nil {
		update.SetIsTemporary(*in.IsTemporary)
	}

	// 更新管理字段
	if in.Status != nil {
		update.SetStatus(uint8(*in.Status))
	}
	if in.Metadata != nil {
		update.SetMetadata(*in.Metadata)
	}
	if len(in.Tags) > 0 {
		// Tags 序列化为 JSON 字符串
		if tagsJSON, err := json.Marshal(in.Tags); err == nil {
			update.SetTags(string(tagsJSON))
		}
	}
	if in.UsageCount != nil {
		update.SetUsageCount(*in.UsageCount)
	}
	if in.LastUsedAt != nil {
		update.SetLastUsedAt(time.Unix(*in.LastUsedAt, 0))
	}

	// 执行更新
	updatedRule, err := update.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("Update casbin rule failed: %v", err)
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("constraint violation during update")
		}
		return nil, fmt.Errorf("update casbin rule failed: %v", err)
	}

	// 🔥 同步更新到 Casbin 引擎
	err = l.syncUpdateToCasbinEngine(rule, updatedRule)
	if err != nil {
		// 记录警告但不回滚数据库操作
		l.Logger.Errorf("Sync updated rule to Casbin engine failed: %v, rule ID: %d", err, *in.Id)
	}

	l.Logger.Infof("Updated casbin rule successfully, ID: %d", *in.Id)

	return &core.BaseResp{
		Msg: "更新权限规则成功",
	}, nil
}

// syncUpdateToCasbinEngine 将更新的规则同步到 Casbin 引擎
func (l *UpdateCasbinRuleLogic) syncUpdateToCasbinEngine(oldRule, newRule *ent.CasbinRule) error {
	// 检查 EnforcerManager 是否可用
	if l.svcCtx.EnforcerManager == nil {
		return fmt.Errorf("EnforcerManager not initialized")
	}

	// 先移除旧规则
	err := l.removeRuleFromCasbinEngine(oldRule)
	if err != nil {
		l.Logger.Errorf("Remove old rule from Casbin engine failed: %v", err)
		// 继续执行，尝试添加新规则
	}

	// 再添加新规则
	err = l.addRuleToCasbinEngine(newRule)
	if err != nil {
		return fmt.Errorf("add new rule to Casbin engine failed: %v", err)
	}

	return nil
}

// removeRuleFromCasbinEngine 从 Casbin 引擎移除规则
// 🔥 适配 RBAC with Domains 模型
func (l *UpdateCasbinRuleLogic) removeRuleFromCasbinEngine(rule *ent.CasbinRule) error {
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
	}
	return nil
}

// addRuleToCasbinEngine 向 Casbin 引擎添加规则
// 🔥 适配 RBAC with Domains 模型
func (l *UpdateCasbinRuleLogic) addRuleToCasbinEngine(rule *ent.CasbinRule) error {
	// 🔥 根据 Ptype 类型添加到 Casbin 引擎
	switch rule.Ptype {
	case "p":
		// 🔥 新格式：v0=sub, v1=domain, v2=obj, v3=act
		// EnforcerManager.AddPolicy 期望：sub, obj, act（内部自动加domain）
		if rule.V0 != "" && rule.V2 != "" && rule.V3 != "" {
			_, err := l.svcCtx.EnforcerManager.AddPolicy(l.ctx, rule.TenantID, rule.V0, rule.V2, rule.V3)
			return err
		}
	case "g":
		// 🔥 新格式：v0=user, v1=role, v2=domain
		// EnforcerManager.AddGroupingPolicy 期望：user, role（内部自动加domain）
		if rule.V0 != "" && rule.V1 != "" {
			_, err := l.svcCtx.EnforcerManager.AddGroupingPolicy(l.ctx, rule.TenantID, rule.V0, rule.V1)
			return err
		}
	case "g2", "g3", "g4", "g5":
		// 🔥 扩展角色继承：v0=user, v1=role, v2=domain
		if rule.V0 != "" && rule.V1 != "" {
			_, err := l.svcCtx.EnforcerManager.AddNamedGroupingPolicy(l.ctx, rule.TenantID, rule.Ptype, rule.V0, rule.V1)
			return err
		}
	}
	return nil
}

// buildCasbinParams 构建 Casbin 参数
// 🔥 注意：此方法已废弃，保留仅为兼容性
// 新的实现直接在 addRuleToCasbinEngine 和 removeRuleFromCasbinEngine 中处理
func (l *UpdateCasbinRuleLogic) buildCasbinParams(rule *ent.CasbinRule) []string {
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
