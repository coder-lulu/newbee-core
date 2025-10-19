package casbin

import (
	"context"
	"fmt"
	"time"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/casbinrule"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserPermissionSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionSummaryLogic {
	return &GetUserPermissionSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 权限查询
func (l *GetUserPermissionSummaryLogic) GetUserPermissionSummary(in *core.GetUserPermissionSummaryReq) (*core.GetUserPermissionSummaryResp, error) {
	// 验证请求参数
	if in.UserId == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)

	// 构建查询 - 必须包含租户ID过滤
	query := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.V0EQ(in.UserId), // 主体为用户ID
			casbinrule.StatusEQ(1),     // 启用状态
		)

	// 按服务过滤
	if in.ServiceName != nil && *in.ServiceName != "" {
		query = query.Where(casbinrule.ServiceNameEQ(*in.ServiceName))
	}

	// 时间过滤 - 只查询当前有效的规则
	now := time.Now()
	query = query.Where(
		casbinrule.Or(
			casbinrule.EffectiveFromLTE(now),
			casbinrule.EffectiveFromIsNil(),
		),
		casbinrule.Or(
			casbinrule.EffectiveToGTE(now),
			casbinrule.EffectiveToIsNil(),
		),
	)

	// 查询直接权限
	directRules, err := query.All(l.ctx)
	if err != nil {
		l.Logger.Errorf("Query direct permissions failed: %v", err)
		return nil, fmt.Errorf("query direct permissions failed: %v", err)
	}

	var allRules = directRules
	var inheritedRules []*ent.CasbinRule

	// 如果需要包含继承权限，查询角色权限
	if in.IncludeInherited != nil && *in.IncludeInherited {
		inheritedRules, err = l.getInheritedPermissions(in.UserId, in.ServiceName)
		if err != nil {
			l.Logger.Errorf("Query inherited permissions failed: %v", err)
			// 不返回错误，只记录日志
		} else {
			allRules = append(allRules, inheritedRules...)
		}
	}

	// 整理权限数据
	permissionMap := make(map[string]*core.PermissionSummary)
	for _, rule := range allRules {
		if rule.V1 == "" || rule.V2 == "" {
			continue // 跳过不完整的规则
		}

		resource := rule.V1
		action := rule.V2

		// 判断是直接权限还是继承权限
		source := "direct"
		for _, inherited := range inheritedRules {
			if rule.ID == inherited.ID {
				source = "inherited"
				break
			}
		}

		// 按资源分组
		key := fmt.Sprintf("%s_%s", resource, source)
		if perm, exists := permissionMap[key]; exists {
			// 添加操作到已有资源
			perm.Actions = l.appendUniqueAction(perm.Actions, action)
		} else {
			// 创建新的权限摘要
			ruleID := fmt.Sprintf("%d", rule.ID)
			permissionMap[key] = &core.PermissionSummary{
				Resource: resource,
				Actions:  []string{action},
				Source:   source,
				RuleId:   &ruleID,
			}
		}
	}

	// 转换为列表
	var permissions []*core.PermissionSummary
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	l.Logger.Infof("Get user permission summary completed: user=%s, total=%d, direct=%d, inherited=%d",
		in.UserId, len(permissions), len(directRules), len(inheritedRules))

	return &core.GetUserPermissionSummaryResp{
		UserId:      in.UserId,
		Permissions: permissions,
		TotalCount:  int32(len(permissions)),
	}, nil
}

// getInheritedPermissions 获取继承权限（通过角色继承）
func (l *GetUserPermissionSummaryLogic) getInheritedPermissions(userID string, serviceName *string) ([]*ent.CasbinRule, error) {
	// 🔥 获取租户ID - 确保多租户隔离安全
	tenantID := tenantctx.GetTenantIDFromCtx(l.ctx)

	// 查询用户的角色关系（g 类型的规则） - 必须包含租户ID过滤
	roleQuery := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.PtypeEQ("g"),
			casbinrule.V0EQ(userID), // 用户
			casbinrule.StatusEQ(1),
		)

	if serviceName != nil && *serviceName != "" {
		roleQuery = roleQuery.Where(casbinrule.ServiceNameEQ(*serviceName))
	}

	userRoles, err := roleQuery.All(l.ctx)
	if err != nil {
		return nil, err
	}

	if len(userRoles) == 0 {
		return []*ent.CasbinRule{}, nil
	}

	// 收集所有角色
	var roles []string
	for _, roleRule := range userRoles {
		if roleRule.V1 != "" {
			roles = append(roles, roleRule.V1)
		}
	}

	if len(roles) == 0 {
		return []*ent.CasbinRule{}, nil
	}

	// 查询角色的权限（p 类型的规则） - 必须包含租户ID过滤
	rolePermQuery := l.svcCtx.DB.CasbinRule.Query().
		Where(
			casbinrule.TenantIDEQ(tenantID), // 🔥 租户隔离
			casbinrule.PtypeEQ("p"),
			casbinrule.V0In(roles...),
			casbinrule.StatusEQ(1),
		)

	if serviceName != nil && *serviceName != "" {
		rolePermQuery = rolePermQuery.Where(casbinrule.ServiceNameEQ(*serviceName))
	}

	// 时间过滤
	now := time.Now()
	rolePermQuery = rolePermQuery.Where(
		casbinrule.Or(
			casbinrule.EffectiveFromLTE(now),
			casbinrule.EffectiveFromIsNil(),
		),
		casbinrule.Or(
			casbinrule.EffectiveToGTE(now),
			casbinrule.EffectiveToIsNil(),
		),
	)

	return rolePermQuery.All(l.ctx)
}

// appendUniqueAction 添加唯一操作
func (l *GetUserPermissionSummaryLogic) appendUniqueAction(actions []string, newAction string) []string {
	for _, action := range actions {
		if action == newAction {
			return actions // 已存在，不重复添加
		}
	}
	return append(actions, newAction)
}
