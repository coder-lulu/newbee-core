package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLogic {
	return &DeleteTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTenantLogic) DeleteTenant(req *types.IDsReq) (resp *types.BaseMsgResp, err error) {
	// 调用RPC服务删除租户
	result, err := l.svcCtx.CoreRpc.DeleteTenant(l.ctx, &core.IDsReq{
		Ids: req.Ids,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, result.Msg)}, nil
}
