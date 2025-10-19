package role

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRoleByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetRoleByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleByIdLogic {
	return &GetRoleByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetRoleByIdLogic) GetRoleById(in *core.IDReq) (*core.RoleInfo, error) {
	result, err := l.svcCtx.DB.Role.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	roleInfo := &core.RoleInfo{
		Id:            &result.ID,
		CreatedAt:     pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:     pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Status:        pointy.GetPointer(uint32(result.Status)),
		Name:          &result.Name,
		Code:          &result.Code,
		DefaultRouter: &result.DefaultRouter,
		Remark:        &result.Remark,
		Sort:          &result.Sort,
		CustomDeptIds: result.CustomDeptIds,
	}

	// 🔥 Phase 3: 从sys_casbin_rules查询数据权限范围
	dataScope, err := getDataScopeFromCasbin(l.ctx, l.svcCtx.DB, result.Code, result.TenantID)
	if err != nil {
		logx.Errorw("Failed to query data scope from casbin",
			logx.Field("role_code", result.Code),
			logx.Field("tenant_id", result.TenantID),
			logx.Field("error", err))
		// 查询失败时设置默认值
		dataScope = 5 // own (最严格的权限)
	}
	roleInfo.DataScope = pointy.GetPointer(dataScope)

	return roleInfo, nil
}
