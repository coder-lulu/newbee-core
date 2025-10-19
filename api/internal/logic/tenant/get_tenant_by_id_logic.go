package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantByIdLogic {
	return &GetTenantByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantByIdLogic) GetTenantById(req *types.IDReq) (resp *types.TenantInfoResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetTenantById(l.ctx, &core.IDReq{
		Id: req.Id,
	})

	if err != nil {
		return nil, err
	}

	return &types.TenantInfoResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  l.svcCtx.Trans.Trans(l.ctx, "common.success"),
		},
		Data: types.TenantInfo{
			BaseIDInfo: types.BaseIDInfo{
				Id:        data.Id,
				CreatedAt: data.CreatedAt,
				UpdatedAt: data.UpdatedAt,
			},
			Status:      data.Status,
			Name:        data.Name,
			Code:        data.Code,
			Description: data.Description,
			ExpiredAt:   data.ExpiredAt,
			Config:      data.Config,
			CreatedBy:   data.CreatedBy,
		},
	}, nil
}
