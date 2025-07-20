package role

import (
	"context"
	"github.com/suyuan32/simple-admin-common/i18n"
	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type AssignRoleDataScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignRoleDataScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRoleDataScopeLogic {
	return &AssignRoleDataScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignRoleDataScopeLogic) AssignRoleDataScope(in *core.RoleDataScopeReq) (*core.BaseResp, error) {
	err := entx.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		_, err := tx.Role.Get(l.ctx, in.Id)
		if err != nil {
			return err
		}

		err = tx.Role.UpdateOneID(in.Id).
			SetNotNilDataScope(pointy.GetStatusPointer(&in.DataScope)).
			SetNotNilCustomDeptIds(in.CustomDeptIds).
			Exec(l.ctx)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
