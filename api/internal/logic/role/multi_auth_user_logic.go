package role

import (
	"context"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MultiAuthUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMultiAuthUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MultiAuthUserLogic {
	return &MultiAuthUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *MultiAuthUserLogic) MultiAuthUser(req *types.RoleAuthReq) (resp *types.BaseMsgResp, err error) {
	result, err := l.svcCtx.CoreRpc.AddAuth(l.ctx, &core.RoleAuthReq{
		RoleId:  req.RoleId,
		UserIds: req.UserIds,
	})
	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, result.Msg)}, nil
}
