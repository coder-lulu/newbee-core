package casbin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateCasbinRuleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateCasbinRuleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateCasbinRuleLogic {
	return &ValidateCasbinRuleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 规则验证
func (l *ValidateCasbinRuleLogic) ValidateCasbinRule(in *core.ValidateCasbinRuleReq) (*core.ValidateCasbinRuleResp, error) {
	if in.Rule == nil {
		return nil, fmt.Errorf("rule is required")
	}

	rule := in.Rule
	var errors []string
	var warnings []string
	var conflicts []string

	// 基本字段验证
	if rule.Ptype == "" {
		errors = append(errors, "ptype is required")
	} else if !l.isValidPtype(rule.Ptype) {
		errors = append(errors, fmt.Sprintf("invalid ptype: %s", rule.Ptype))
	}

	if rule.ServiceName == "" {
		errors = append(errors, "service_name is required")
	}

	// Casbin标准字段验证
	if rule.V0 != nil && *rule.V0 == "" {
		warnings = append(warnings, "v0 (subject) is empty")
	}
	if rule.V1 != nil && *rule.V1 == "" {
		warnings = append(warnings, "v1 (object) is empty")
	}
	if rule.V2 != nil && *rule.V2 == "" {
		warnings = append(warnings, "v2 (action) is empty")
	}

	// 效果验证
	if rule.V3 != nil && *rule.V3 != "" {
		effect := strings.ToLower(*rule.V3)
		if effect != "allow" && effect != "deny" {
			errors = append(errors, fmt.Sprintf("invalid effect: %s, must be 'allow' or 'deny'", *rule.V3))
		}
	}

	// 时间验证
	if rule.EffectiveFrom != nil && rule.EffectiveTo != nil {
		from := time.Unix(*rule.EffectiveFrom, 0)
		to := time.Unix(*rule.EffectiveTo, 0)
		if from.After(to) {
			errors = append(errors, "effective_from must be before effective_to")
		}
	}

	// 审批流程验证
	if rule.RequireApproval != nil && *rule.RequireApproval {
		if rule.ApprovalStatus != nil && *rule.ApprovalStatus != "" {
			if !l.isValidApprovalStatus(*rule.ApprovalStatus) {
				errors = append(errors, fmt.Sprintf("invalid approval_status: %s", *rule.ApprovalStatus))
			}
			if *rule.ApprovalStatus == "approved" && (rule.ApprovedBy == nil || *rule.ApprovedBy == 0) {
				warnings = append(warnings, "approved rule should have approved_by")
			}
		}
	}

	// 状态验证
	if rule.Status != nil && (*rule.Status != 0 && *rule.Status != 1) {
		errors = append(errors, "status must be 0 (disabled) or 1 (enabled)")
	}

	// 检查冲突
	if in.CheckConflicts != nil && *in.CheckConflicts {
		conflictIDs, err := l.checkRuleConflicts(rule)
		if err != nil {
			l.Logger.Errorf("Check rule conflicts failed: %v", err)
			warnings = append(warnings, fmt.Sprintf("failed to check conflicts: %v", err))
		} else {
			conflicts = conflictIDs
		}
	}

	// 验证结果
	isValid := len(errors) == 0

	l.Logger.Infof("Rule validation completed: valid=%t, errors=%d, warnings=%d, conflicts=%d",
		isValid, len(errors), len(warnings), len(conflicts))

	return &core.ValidateCasbinRuleResp{
		Valid:     isValid,
		Errors:    errors,
		Warnings:  warnings,
		Conflicts: conflicts,
	}, nil
}

// isValidPtype 检查ptype是否有效
func (l *ValidateCasbinRuleLogic) isValidPtype(ptype string) bool {
	validPtypes := []string{"p", "g", "g2", "g3", "g4"}
	for _, valid := range validPtypes {
		if ptype == valid {
			return true
		}
	}
	return false
}

// isValidApprovalStatus 检查审批状态是否有效
func (l *ValidateCasbinRuleLogic) isValidApprovalStatus(status string) bool {
	validStatuses := []string{"pending", "approved", "rejected"}
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// checkRuleConflicts 检查规则冲突
func (l *ValidateCasbinRuleLogic) checkRuleConflicts(rule *core.CasbinRuleInfo) ([]string, error) {
	var conflicts []string

	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)

	// 查询相似的规则 - 必须包含租户ID过滤
	query := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.ServiceNameEQ(rule.ServiceName),
			casbinrule.PtypeEQ(rule.Ptype),
			casbinrule.StatusEQ(1), // 只检查启用的规则
		)

	// 排除自身（如果是更新操作）
	if rule.Id != nil && *rule.Id > 0 {
		query = query.Where(casbinrule.IDNEQ(*rule.Id))
	}

	// 添加v0-v2的匹配条件
	if rule.V0 != nil {
		query = query.Where(casbinrule.V0EQ(*rule.V0))
	}
	if rule.V1 != nil {
		query = query.Where(casbinrule.V1EQ(*rule.V1))
	}
	if rule.V2 != nil {
		query = query.Where(casbinrule.V2EQ(*rule.V2))
	}

	existingRules, err := query.All(l.ctx)
	if err != nil {
		return nil, err
	}

	// 检查是否有相同的规则
	for _, existing := range existingRules {
		if l.isRuleConflict(rule, existing) {
			conflicts = append(conflicts, fmt.Sprintf("%d", existing.ID))
		}
	}

	return conflicts, nil
}

// isRuleConflict 检查两个规则是否冲突
func (l *ValidateCasbinRuleLogic) isRuleConflict(newRule *core.CasbinRuleInfo, existingRule *ent.CasbinRule) bool {
	// 检查是否为完全相同的规则
	if newRule.V0 != nil && *newRule.V0 == existingRule.V0 &&
		newRule.V1 != nil && *newRule.V1 == existingRule.V1 &&
		newRule.V2 != nil && *newRule.V2 == existingRule.V2 {

		// 检查是否有相反的效果
		if newRule.V3 != nil && existingRule.V3 != "" {
			newEffect := strings.ToLower(*newRule.V3)
			existingEffect := strings.ToLower(existingRule.V3)
			if (newEffect == "allow" && existingEffect == "deny") ||
				(newEffect == "deny" && existingEffect == "allow") {
				return true // 有冲突
			}
		}

		// 检查是否为重复规则
		return true
	}

	return false
}
