package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitTenantDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitTenantDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitTenantDataLogic {
	return &InitTenantDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitTenantDataLogic) InitTenantData(req *types.IDReq) (resp *types.BaseMsgResp, err error) {
	// 调用RPC服务初始化租户数据
	result, err := l.svcCtx.CoreRpc.InitTenantData(l.ctx, &core.IDReq{
		Id: req.Id,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, result.Msg)}, nil
}
