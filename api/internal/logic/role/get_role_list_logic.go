package role

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/v2/i18n"
)

type GetRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRoleListLogic {
	return &GetRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRoleListLogic) GetRoleList(req *types.RoleListReq) (resp *types.RoleListResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetRoleList(l.ctx,
		&core.RoleListReq{
			Page:     req.Page,
			PageSize: req.PageSize,
			Name:     req.Name,
			Code:     req.Code,
		})
	if err != nil {
		return nil, err
	}
	resp = &types.RoleListResp{}
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, i18n.Success)
	resp.Data.Total = data.GetTotal()

	for _, v := range data.Data {
		resp.Data.Data = append(resp.Data.Data,
			types.RoleInfo{
				BaseIDInfo: types.BaseIDInfo{
					Id:        v.Id,
					CreatedAt: v.CreatedAt,
					UpdatedAt: v.UpdatedAt,
				},
				Trans:         l.svcCtx.Trans.Trans(l.ctx, *v.Name),
				Status:        v.Status,
				Name:          v.Name,
				Code:          v.Code,
				DefaultRouter: v.DefaultRouter,
				Remark:        v.Remark,
				Sort:          v.Sort,
				DataScope:     v.DataScope,
				CustomDeptIds: v.CustomDeptIds,
			})
	}
	return resp, nil
}
