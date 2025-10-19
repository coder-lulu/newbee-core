package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantByCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantByCodeLogic {
	return &GetTenantByCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantByCodeLogic) GetTenantByCode(req *types.TenantCodeReq) (resp *types.TenantInfoResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetTenantByCode(l.ctx, &core.TenantCodeReq{
		Code: req.Code,
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
