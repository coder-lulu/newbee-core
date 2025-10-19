package role

import (
	"context"
	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangeRoleStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangeRoleStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeRoleStatusLogic {
	return &ChangeRoleStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *ChangeRoleStatusLogic) ChangeRoleStatus(req *types.RoleChangeStatusReq) (resp *types.BaseMsgResp, err error) {
	data, err := l.svcCtx.CoreRpc.ChangeRoleStatus(l.ctx,
		&core.RoleStatusChangeParam{
			Id:     *req.Id,
			Status: *req.Status,
		})
	if err != nil {
		return nil, err
	}

	if req.Status != nil && uint8(*req.Status) == common.StatusBanned {
		_, err := l.svcCtx.CoreRpc.GetRoleById(l.ctx, &core.IDReq{Id: *req.Id})
		if err != nil {
			return nil, err
		}

	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, data.Msg)}, nil
}
