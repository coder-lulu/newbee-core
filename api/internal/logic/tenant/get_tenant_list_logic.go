package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantListLogic) GetTenantList(req *types.TenantListReq) (resp *types.TenantListResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetTenantList(l.ctx, &core.TenantListReq{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Name:      req.Name,
		Code:      req.Code,
		Status:    req.Status,
		CreatedBy: req.CreatedBy,
	})

	if err != nil {
		return nil, err
	}

	resp = &types.TenantListResp{}
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, "common.success")
	resp.Data.Total = data.Total
	list := make([]types.TenantInfo, 0)

	for _, v := range data.Data {
		list = append(list, types.TenantInfo{
			BaseIDInfo: types.BaseIDInfo{
				Id:        v.Id,
				CreatedAt: v.CreatedAt,
				UpdatedAt: v.UpdatedAt,
			},
			Status:      v.Status,
			Name:        v.Name,
			Code:        v.Code,
			Description: v.Description,
			ExpiredAt:   v.ExpiredAt,
			Config:      v.Config,
			CreatedBy:   v.CreatedBy,
		})
	}

	resp.Data.Data = list
	return resp, nil
}
