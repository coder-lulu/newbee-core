package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantLogic) UpdateTenant(req *types.TenantInfo) (resp *types.BaseMsgResp, err error) {
	data, err := l.svcCtx.CoreRpc.UpdateTenant(l.ctx, &core.TenantInfo{
		Id:          req.Id,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		ExpiredAt:   req.ExpiredAt,
		Config:      req.Config,
		CreatedBy:   req.CreatedBy,
		Status:      req.Status,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{
		Code: 0,
		Msg:  l.svcCtx.Trans.Trans(l.ctx, data.Msg),
	}, nil
}
