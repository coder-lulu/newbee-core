package casbin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCasbinRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCasbinRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCasbinRuleLogic {
	return &CreateCasbinRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 权限规则管理
func (l *CreateCasbinRuleLogic) CreateCasbinRule(in *core.CasbinRuleInfo) (*core.BaseIDResp, error) {
	// 验证必需字段
	if in.Ptype == "" {
		return nil, fmt.Errorf("ptype is required")
	}
	if in.ServiceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}

	// 构建创建器
	create := l.svcCtx.DB.CasbinRule.Create().
		SetPtype(in.Ptype).
		SetServiceName(in.ServiceName)

	// 设置可选的 Casbin 标准字段
	if in.V0 != nil {
		create.SetV0(*in.V0)
	}
	if in.V1 != nil {
		create.SetV1(*in.V1)
	}
	if in.V2 != nil {
		create.SetV2(*in.V2)
	}
	if in.V3 != nil {
		create.SetV3(*in.V3)
	}
	if in.V4 != nil {
		create.SetV4(*in.V4)
	}
	if in.V5 != nil {
		create.SetV5(*in.V5)
	}

	// 设置业务扩展字段
	if in.RuleName != nil {
		create.SetRuleName(*in.RuleName)
	}
	if in.Description != nil {
		create.SetDescription(*in.Description)
	}
	if in.Category != nil {
		create.SetCategory(*in.Category)
	}
	if in.Version != nil {
		create.SetVersion(*in.Version)
	}

	// 设置审批流程字段
	if in.RequireApproval != nil {
		create.SetRequireApproval(*in.RequireApproval)
	}
	if in.ApprovalStatus != nil {
		create.SetApprovalStatus(casbinrule.ApprovalStatus(*in.ApprovalStatus))
	}
	if in.ApprovedBy != nil {
		create.SetApprovedBy(*in.ApprovedBy)
	}
	if in.ApprovedAt != nil {
		create.SetApprovedAt(time.Unix(*in.ApprovedAt, 0))
	}

	// 设置时间控制字段
	if in.EffectiveFrom != nil {
		create.SetEffectiveFrom(time.Unix(*in.EffectiveFrom, 0))
	}
	if in.EffectiveTo != nil {
		create.SetEffectiveTo(time.Unix(*in.EffectiveTo, 0))
	}
	if in.IsTemporary != nil {
		create.SetIsTemporary(*in.IsTemporary)
	}

	// 设置管理字段
	if in.Status != nil {
		create.SetStatus(uint8(*in.Status))
	} else {
		create.SetStatus(1) // 默认启用
	}
	if in.Metadata != nil {
		create.SetMetadata(*in.Metadata)
	}
	if len(in.Tags) > 0 {
		// 将字符串数组序列化为JSON字符串
		tagsJSON, _ := json.Marshal(in.Tags)
		create.SetTags(string(tagsJSON))
	}

	// 初始化统计字段
	create.SetUsageCount(0)
	create.SetLastUsedAt(time.Now())

	// 执行创建
	result, err := create.Save(l.ctx)
	if err != nil {
		l.Logger.Errorf("Create casbin rule failed: %v", err)
		if ent.IsConstraintError(err) {
			return nil, fmt.Errorf("rule already exists or constraint violation")
		}
		return nil, fmt.Errorf("create casbin rule failed: %v", err)
	}

	// 🔥 同步到 Casbin 引擎
	err = l.syncToCasbinEngine(result)
	if err != nil {
		// 记录警告但不回滚数据库操作
		l.Logger.Errorf("Sync rule to Casbin engine failed: %v, rule ID: %d", err, result.ID)
		// 可以考虑在这里添加重试机制或异步同步
	}

	l.Logger.Infof("Created casbin rule successfully, ID: %d, ptype: %s, service: %s",
		result.ID, result.Ptype, result.ServiceName)

	return &core.BaseIDResp{
		Id:  result.ID,
		Msg: "创建权限规则成功",
	}, nil
}

// syncToCasbinEngine 将新创建的规则同步到 Casbin 引擎
// 🔥 适配 RBAC with Domains 模型
func (l *CreateCasbinRuleLogic) syncToCasbinEngine(rule *ent.CasbinRule) error {
	// 检查 EnforcerManager 是否可用
	if l.svcCtx.EnforcerManager == nil {
		return fmt.Errorf("EnforcerManager not initialized")
	}

	// 🔥 根据 Ptype 类型添加到 Casbin 引擎
	// 注意：EnforcerManager 内部会自动处理 domain，所以这里不需要传递 v1 (domain)
	switch rule.Ptype {
	case "p":
		// 🔥 新格式：v0=sub, v1=domain, v2=obj, v3=act
		// EnforcerManager.AddPolicy 期望：sub, obj, act（内部自动加domain）
		if rule.V0 != "" && rule.V2 != "" && rule.V3 != "" {
			_, err := l.svcCtx.EnforcerManager.AddPolicy(l.ctx, rule.TenantID, rule.V0, rule.V2, rule.V3)
			return err
		}
		l.Logger.Errorf("Policy rule incomplete: v0=%s, v2=%s, v3=%s", rule.V0, rule.V2, rule.V3)

	case "g":
		// 🔥 新格式：v0=user, v1=role, v2=domain
		// EnforcerManager.AddGroupingPolicy 期望：user, role（内部自动加domain）
		if rule.V0 != "" && rule.V1 != "" {
			_, err := l.svcCtx.EnforcerManager.AddGroupingPolicy(l.ctx, rule.TenantID, rule.V0, rule.V1)
			return err
		}
		l.Logger.Errorf("Grouping rule incomplete: v0=%s, v1=%s", rule.V0, rule.V1)

	case "g2", "g3", "g4", "g5":
		// 🔥 扩展角色继承：v0=user, v1=role, v2=domain
		if rule.V0 != "" && rule.V1 != "" {
			_, err := l.svcCtx.EnforcerManager.AddNamedGroupingPolicy(l.ctx, rule.TenantID, rule.Ptype, rule.V0, rule.V1)
			return err
		}
		l.Logger.Errorf("Named grouping rule incomplete: v0=%s, v1=%s", rule.V0, rule.V1)

	default:
		l.Logger.Infof("Unknown ptype: %s, skipping Casbin engine sync", rule.Ptype)
	}

	return nil
}
