package casbin

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCasbinRuleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCasbinRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCasbinRuleListLogic {
	return &GetCasbinRuleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCasbinRuleListLogic) GetCasbinRuleList(in *core.CasbinRuleListReq) (*core.CasbinRuleListResp, error) {
	// 🔥 获取租户ID - 支持SystemContext加载所有租户规则
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)
	isSystemContext := hooks.IsSystemContext(l.ctx)

	// 构建查询
	query := l.svcCtx.DB.CasbinRule.Query()

	// ⚠️ 租户隔离策略：
	// - SystemContext: 不添加租户过滤（用于API服务初始化加载所有租户规则）
	// - 普通Context: 强制按租户ID过滤（安全隔离）
	if !isSystemContext {
		query = query.Where(casbinrule.TenantIDEQ(tenantID))
		l.Logger.Infow("GetCasbinRuleList with tenant filter",
			logx.Field("tenant_id", tenantID))
	} else {
		l.Logger.Infow("GetCasbinRuleList with SystemContext, loading all tenants' rules")
	}

	// 应用过滤条件
	if in.ServiceName != nil && *in.ServiceName != "" {
		query = query.Where(casbinrule.ServiceNameEQ(*in.ServiceName))
	}
	if in.Ptype != nil && *in.Ptype != "" {
		query = query.Where(casbinrule.PtypeEQ(*in.Ptype))
	}
	if in.V0 != nil && *in.V0 != "" {
		query = query.Where(casbinrule.V0EQ(*in.V0))
	}
	if in.V1 != nil && *in.V1 != "" {
		query = query.Where(casbinrule.V1EQ(*in.V1))
	}
	if in.Status != nil {
		query = query.Where(casbinrule.StatusEQ(uint8(*in.Status)))
	}
	if in.ApprovalStatus != nil && *in.ApprovalStatus != "" {
		query = query.Where(casbinrule.ApprovalStatusEQ(casbinrule.ApprovalStatus(*in.ApprovalStatus)))
	}
	if in.Category != nil && *in.Category != "" {
		query = query.Where(casbinrule.CategoryEQ(*in.Category))
	}
	if in.IsTemporary != nil {
		query = query.Where(casbinrule.IsTemporaryEQ(*in.IsTemporary))
	}

	// 时间范围过滤
	if in.EffectiveFromStart != nil {
		query = query.Where(casbinrule.EffectiveFromGTE(time.Unix(*in.EffectiveFromStart, 0)))
	}
	if in.EffectiveFromEnd != nil {
		query = query.Where(casbinrule.EffectiveFromLTE(time.Unix(*in.EffectiveFromEnd, 0)))
	}

	// 关键词搜索
	if in.Keyword != nil && *in.Keyword != "" {
		keyword := strings.TrimSpace(*in.Keyword)
		query = query.Where(
			casbinrule.Or(
				casbinrule.RuleNameContainsFold(keyword),
				casbinrule.DescriptionContainsFold(keyword),
			),
		)
	}

	// 统计总数
	total, err := query.Count(l.ctx)
	if err != nil {
		l.Logger.Errorf("Count casbin rules failed: %v", err)
		return nil, err
	}

	// 排序
	orderBy := "created_at"
	orderDirection := "desc"
	if in.OrderBy != nil && *in.OrderBy != "" {
		orderBy = *in.OrderBy
	}
	if in.OrderDirection != nil && *in.OrderDirection != "" {
		orderDirection = *in.OrderDirection
	}

	switch orderBy {
	case "id":
		if orderDirection == "asc" {
			query = query.Order(ent.Asc(casbinrule.FieldID))
		} else {
			query = query.Order(ent.Desc(casbinrule.FieldID))
		}
	case "usage_count":
		if orderDirection == "asc" {
			query = query.Order(ent.Asc(casbinrule.FieldUsageCount))
		} else {
			query = query.Order(ent.Desc(casbinrule.FieldUsageCount))
		}
	case "last_used_at":
		if orderDirection == "asc" {
			query = query.Order(ent.Asc(casbinrule.FieldLastUsedAt))
		} else {
			query = query.Order(ent.Desc(casbinrule.FieldLastUsedAt))
		}
	default: // created_at
		if orderDirection == "asc" {
			query = query.Order(ent.Asc(casbinrule.FieldCreatedAt))
		} else {
			query = query.Order(ent.Desc(casbinrule.FieldCreatedAt))
		}
	}

	// 分页
	offset := int((in.Page - 1) * in.PageSize)
	limit := int(in.PageSize)

	// 查询数据
	rules, err := query.Offset(offset).Limit(limit).All(l.ctx)
	if err != nil {
		l.Logger.Errorf("Query casbin rules failed: %v", err)
		return nil, err
	}

	// 转换数据格式
	var data []*core.CasbinRuleInfo
	for _, rule := range rules {
		ruleInfo := l.convertToRuleInfo(rule)
		data = append(data, ruleInfo)
	}

	return &core.CasbinRuleListResp{
		Total: uint64(total),
		Data:  data,
	}, nil
}

// convertToRuleInfo 转换ent模型为protobuf模型
func (l *GetCasbinRuleListLogic) convertToRuleInfo(rule *ent.CasbinRule) *core.CasbinRuleInfo {
	ruleInfo := &core.CasbinRuleInfo{
		Ptype:       rule.Ptype,
		ServiceName: rule.ServiceName,
	}

	// 设置ID和时间戳
	if rule.ID > 0 {
		ruleInfo.Id = &rule.ID
	}
	if !rule.CreatedAt.IsZero() {
		createdAt := rule.CreatedAt.Unix()
		ruleInfo.CreatedAt = &createdAt
	}
	if !rule.UpdatedAt.IsZero() {
		updatedAt := rule.UpdatedAt.Unix()
		ruleInfo.UpdatedAt = &updatedAt
	}
	if rule.TenantID > 0 {
		ruleInfo.TenantId = &rule.TenantID
	}

	// 设置可选的Casbin标准字段
	if rule.V0 != "" {
		ruleInfo.V0 = &rule.V0
	}
	if rule.V1 != "" {
		ruleInfo.V1 = &rule.V1
	}
	if rule.V2 != "" {
		ruleInfo.V2 = &rule.V2
	}
	if rule.V3 != "" {
		ruleInfo.V3 = &rule.V3
	}
	if rule.V4 != "" {
		ruleInfo.V4 = &rule.V4
	}
	if rule.V5 != "" {
		ruleInfo.V5 = &rule.V5
	}

	// 设置业务扩展字段
	if rule.RuleName != "" {
		ruleInfo.RuleName = &rule.RuleName
	}
	if rule.Description != "" {
		ruleInfo.Description = &rule.Description
	}
	if rule.Category != "" {
		ruleInfo.Category = &rule.Category
	}
	if rule.Version != "" {
		ruleInfo.Version = &rule.Version
	}

	// 设置审批流程字段
	ruleInfo.RequireApproval = &rule.RequireApproval
	if rule.ApprovalStatus != "" {
		approvalStatusStr := string(rule.ApprovalStatus)
		ruleInfo.ApprovalStatus = &approvalStatusStr
	}
	if rule.ApprovedBy > 0 {
		ruleInfo.ApprovedBy = &rule.ApprovedBy
	}
	if !rule.ApprovedAt.IsZero() {
		approvedAt := rule.ApprovedAt.Unix()
		ruleInfo.ApprovedAt = &approvedAt
	}

	// 设置时间控制字段
	if !rule.EffectiveFrom.IsZero() {
		effectiveFrom := rule.EffectiveFrom.Unix()
		ruleInfo.EffectiveFrom = &effectiveFrom
	}
	if !rule.EffectiveTo.IsZero() {
		effectiveTo := rule.EffectiveTo.Unix()
		ruleInfo.EffectiveTo = &effectiveTo
	}
	ruleInfo.IsTemporary = &rule.IsTemporary

	// 设置管理字段
	if rule.Status > 0 {
		status := uint32(rule.Status)
		ruleInfo.Status = &status
	}
	if rule.Metadata != "" {
		ruleInfo.Metadata = &rule.Metadata
	}
	if rule.Tags != "" {
		// Tags 从 JSON 字符串解析为 []string
		var tags []string
		if err := json.Unmarshal([]byte(rule.Tags), &tags); err == nil {
			ruleInfo.Tags = tags
		}
	}
	if rule.UsageCount > 0 {
		ruleInfo.UsageCount = &rule.UsageCount
	}
	if !rule.LastUsedAt.IsZero() {
		lastUsedAt := rule.LastUsedAt.Unix()
		ruleInfo.LastUsedAt = &lastUsedAt
	}

	return ruleInfo
}
