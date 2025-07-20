package role

import (
	"context"
	"fmt"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRoleDataScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRoleDataScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleDataScopeLogic {
	return &AssignRoleDataScopeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *AssignRoleDataScopeLogic) AssignRoleDataScope(in *types.RoleDataScopeReq) (resp *types.BaseMsgResp, err error) {
	fmt.Println(in.DataScope)
	data, err := l.svcCtx.CoreRpc.AssignRoleDataScope(l.ctx,
		&core.RoleDataScopeReq{
			Id:            *in.Id,
			DataScope:     in.DataScope,
			CustomDeptIds: in.CustomDeptIds,
		})
	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, data.Msg)}, nil
}
