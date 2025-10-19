package role

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleListLogic {
	return &GetRoleListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRoleListLogic) GetRoleList(in *core.RoleListReq) (*core.RoleListResp, error) {
	var predicates []predicate.Role
	if in.Name != nil && *in.Name != "" {
		predicates = append(predicates, role.NameContains(*in.Name))
	}
	if in.Code != nil && *in.Code != "" {
		predicates = append(predicates, role.CodeEQ(*in.Code))
	}
	if in.DefaultRouter != nil && *in.DefaultRouter != "" {
		predicates = append(predicates, role.DefaultRouterContains(*in.DefaultRouter))
	}

	result, err := l.svcCtx.DB.Role.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize, func(pager *ent.RolePager) {
		pager.Order = ent.Asc(role.FieldSort)
	})
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.RoleListResp{}
	resp.Total = result.PageDetails.Total

	// 🔥 Phase 3: 从sys_casbin_rules查询数据权限范围
	for _, v := range result.List {
		roleInfo := &core.RoleInfo{
			Id:            &v.ID,
			CreatedAt:     pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:     pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:        pointy.GetPointer(uint32(v.Status)),
			Name:          &v.Name,
			Code:          &v.Code,
			DefaultRouter: &v.DefaultRouter,
			Remark:        &v.Remark,
			Sort:          &v.Sort,
			CustomDeptIds: v.CustomDeptIds,
		}

		// 查询数据权限范围（从sys_casbin_rules表，ptype='d'）
		dataScope, err := getDataScopeFromCasbin(l.ctx, l.svcCtx.DB, v.Code, v.TenantID)
		if err != nil {
			logx.Errorw("Failed to query data scope from casbin",
				logx.Field("role_code", v.Code),
				logx.Field("tenant_id", v.TenantID),
				logx.Field("error", err))
			// 查询失败时设置默认值
			dataScope = 5 // own (最严格的权限)
		}
		roleInfo.DataScope = pointy.GetPointer(dataScope)

		resp.Data = append(resp.Data, roleInfo)
	}

	return resp, nil
}
