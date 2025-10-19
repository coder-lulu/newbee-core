package authority

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrUpdateApiAuthorityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrUpdateApiAuthorityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrUpdateApiAuthorityLogic {
	return &CreateOrUpdateApiAuthorityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrUpdateApiAuthorityLogic) CreateOrUpdateApiAuthority(req *types.CreateOrUpdateApiAuthorityReq) (resp *types.BaseMsgResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetRoleById(l.ctx, &core.IDReq{Id: req.RoleId})
	if err != nil {
		return nil, err
	}

	// 从上下文获取租户ID
	tenantID := l.svcCtx.ContextManager.GetTenantID(l.ctx)
	if tenantID == "" {
		l.Logger.Errorw("failed to get tenant ID from context", logx.Field("roleId", req.RoleId))
		return nil, errorx.NewInternalError("tenant ID not found in context")
	}

	tenantUint, convErr := strconv.ParseUint(tenantID, 10, 64)
	if convErr != nil {
		l.Logger.Errorw("invalid tenant id format", logx.Field("tenantID", tenantID), logx.Field("error", convErr.Error()))
		return nil, errorx.NewInternalError("invalid tenant ID in context")
	}

	roleCode := ""
	if data.Code != nil {
		roleCode = *data.Code
	}
	if roleCode == "" {
		return nil, errorx.NewInternalError("role code is empty")
	}

	const serviceName = "core"
	existingRules, err := l.listExistingPolicies(roleCode, tenantID, serviceName)
	if err != nil {
		l.Logger.Errorw("failed to list existing casbin policies", logx.Field("roleCode", roleCode), logx.Field("error", err.Error()))
		return nil, errorx.NewInternalError("failed to query existing policies")
	}

	toAdd, toDelete := l.diffPolicies(existingRules, req.Data)

	if len(toDelete) > 0 {
		_, err = l.svcCtx.CoreRpc.BatchDeleteCasbinRules(l.ctx, &core.IDsReq{Ids: toDelete})
		if err != nil {
			l.Logger.Errorw("failed to delete stale policies via RPC", logx.Field("roleCode", roleCode), logx.Field("error", err.Error()))
			return nil, errorx.NewInternalError("failed to delete existing policies")
		}
	}

	if len(toAdd) > 0 {
		newRules := make([]*core.CasbinRuleInfo, 0, len(toAdd))
		for _, item := range toAdd {
			method := strings.ToUpper(item.Method)
			path := item.Path
			effect := "allow"

			newRules = append(newRules, &core.CasbinRuleInfo{
				Ptype:       "p",
				ServiceName: serviceName,
				V0:          pointy.GetPointer(roleCode),
				V1:          pointy.GetPointer(tenantID),
				V2:          pointy.GetPointer(path),
				V3:          pointy.GetPointer(method),
				V4:          pointy.GetPointer(effect),
				TenantId:    pointy.GetPointer(tenantUint),
				Status:      pointy.GetPointer(uint32(1)),
			})
		}

		_, err = l.svcCtx.CoreRpc.BatchCreateCasbinRules(l.ctx, &core.BatchCreateCasbinRulesReq{Rules: newRules})
		if err != nil {
			l.Logger.Errorw("failed to create casbin policies via RPC",
				logx.Field("roleCode", roleCode),
				logx.Field("tenantID", tenantID),
				logx.Field("error", err.Error()))
			return nil, errorx.NewInternalError("failed to create policies")
		}
	}

	// 主动刷新本地策略缓存，避免等待Redis事件
	if err := l.svcCtx.Casbin.LoadPolicy(); err != nil {
		l.Logger.Errorw("failed to reload casbin policy after update", logx.Field("error", err.Error()))
	}

	l.Logger.Infow("successfully synchronized API policies via RPC",
		logx.Field("roleCode", roleCode),
		logx.Field("tenantID", tenantID),
		logx.Field("added", len(toAdd)),
		logx.Field("deleted", len(toDelete)))

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, i18n.UpdateSuccess)}, nil
}

// listExistingPolicies 分页查询指定角色和租户的现有API策略
func (l *CreateOrUpdateApiAuthorityLogic) listExistingPolicies(roleCode, tenantID, serviceName string) ([]*core.CasbinRuleInfo, error) {
	const pageSize uint64 = 500
	page := uint64(1)
	var result []*core.CasbinRuleInfo

	for {
		req := &core.CasbinRuleListReq{
			Page:        page,
			PageSize:    pageSize,
			ServiceName: pointy.GetPointer(serviceName),
			Ptype:       pointy.GetPointer("p"),
			V0:          pointy.GetPointer(roleCode),
			V1:          pointy.GetPointer(tenantID),
		}

		resp, err := l.svcCtx.CoreRpc.GetCasbinRuleList(l.ctx, req)
		if err != nil {
			return nil, err
		}

		result = append(result, resp.Data...)

		if uint64(len(resp.Data)) < pageSize {
			break
		}

		page++
	}

	return result, nil
}

// diffPolicies 计算需要新增和删除的策略集合
func (l *CreateOrUpdateApiAuthorityLogic) diffPolicies(existing []*core.CasbinRuleInfo, desired []types.ApiAuthorityInfo) (toAdd []types.ApiAuthorityInfo, toDelete []uint64) {
	existingByKey := make(map[string]*core.CasbinRuleInfo, len(existing))
	for _, rule := range existing {
		if rule == nil {
			continue
		}
		key := normalizePolicyKey(rule.GetV2(), rule.GetV3())
		existingByKey[key] = rule
	}

	desiredKeys := make(map[string]struct{}, len(desired))
	for _, item := range desired {
		key := normalizePolicyKey(item.Path, item.Method)
		desiredKeys[key] = struct{}{}

		if _, exists := existingByKey[key]; !exists {
			toAdd = append(toAdd, item)
		}
	}

	for key, rule := range existingByKey {
		if _, keep := desiredKeys[key]; !keep && rule.Id != nil {
			toDelete = append(toDelete, *rule.Id)
		}
	}

	return toAdd, toDelete
}

// normalizePolicyKey 统一策略键格式: METHOD::PATH
func normalizePolicyKey(path, method string) string {
	return strings.ToUpper(method) + "::" + path
}
