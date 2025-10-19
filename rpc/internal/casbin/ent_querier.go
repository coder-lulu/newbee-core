package casbin

import (
	"context"

	commontypes "github.com/coder-lulu/newbee-common/casbin/types"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
)

// EntCasbinRuleQuerier RPC服务端的Casbin规则查询器
// 直接使用ent client查询数据库
type EntCasbinRuleQuerier struct {
	db *ent.Client
}

// NewEntCasbinRuleQuerier 创建查询器
func NewEntCasbinRuleQuerier(db *ent.Client) *EntCasbinRuleQuerier {
	return &EntCasbinRuleQuerier{
		db: db,
	}
}

// QueryCasbinRules 实现CasbinRuleQuerier接口
// 查询指定租户的所有Casbin规则
func (q *EntCasbinRuleQuerier) QueryCasbinRules(ctx context.Context, tenantID uint64) ([]commontypes.CasbinRuleEntity, error) {
	// 查询数据库中的规则
	rules, err := q.db.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID),
			casbinrule.StatusEQ(1), // 只查询启用的规则
		).
		All(ctx)

	if err != nil {
		return nil, err
	}

	// 转换为通用接口
	result := make([]commontypes.CasbinRuleEntity, len(rules))
	for i, rule := range rules {
		result[i] = &EntCasbinRuleWrapper{rule: rule}
	}

	return result, nil
}

// EntCasbinRuleWrapper 包装ent.CasbinRule以实现CasbinRuleEntity接口
type EntCasbinRuleWrapper struct {
	rule *ent.CasbinRule
}

// 实现CasbinRuleEntity接口的所有方法

func (w *EntCasbinRuleWrapper) GetID() uint64 {
	return w.rule.ID
}

func (w *EntCasbinRuleWrapper) GetPtype() string {
	return w.rule.Ptype
}

func (w *EntCasbinRuleWrapper) GetV0() string {
	return w.rule.V0
}

func (w *EntCasbinRuleWrapper) GetV1() string {
	return w.rule.V1
}

func (w *EntCasbinRuleWrapper) GetV2() string {
	return w.rule.V2
}

func (w *EntCasbinRuleWrapper) GetV3() string {
	return w.rule.V3
}

func (w *EntCasbinRuleWrapper) GetV4() string {
	return w.rule.V4
}

func (w *EntCasbinRuleWrapper) GetV5() string {
	return w.rule.V5
}

func (w *EntCasbinRuleWrapper) GetTenantID() uint64 {
	return w.rule.TenantID
}

func (w *EntCasbinRuleWrapper) GetStatus() uint8 {
	return w.rule.Status
}

func (w *EntCasbinRuleWrapper) GetRequireApproval() bool {
	return w.rule.RequireApproval
}

func (w *EntCasbinRuleWrapper) GetApprovalStatus() string {
	return string(w.rule.ApprovalStatus)
}

func (w *EntCasbinRuleWrapper) HasEffectiveFrom() bool {
	return !w.rule.EffectiveFrom.IsZero()
}

func (w *EntCasbinRuleWrapper) HasEffectiveTo() bool {
	return !w.rule.EffectiveTo.IsZero()
}
