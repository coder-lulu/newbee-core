package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantLogic {
	return &InitTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitTenantLogic) InitTenant(req *types.TenantInitReq) (resp *types.BaseMsgResp, err error) {
	data, err := l.svcCtx.CoreRpc.InitTenant(l.ctx, &core.TenantInitReq{
		TenantId:      req.TenantId,
		AdminUsername: req.AdminUsername,
		AdminPassword: req.AdminPassword,
		AdminEmail:    req.AdminEmail,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{
		Code: 0,
		Msg:  l.svcCtx.Trans.Trans(l.ctx, data.Msg),
	}, nil
}
