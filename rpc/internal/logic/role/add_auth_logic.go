package role

import (
	"context"

	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	uuid "github.com/gofrs/uuid/v5"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddAuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddAuthLogic {
	return &AddAuthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddAuthLogic) AddAuth(in *core.RoleAuthReq) (*core.BaseResp, error) {
	err := entx.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		_, err := l.svcCtx.DB.Role.Query().Where(role.IDEQ(in.RoleId)).First(l.ctx)
		if err != nil {
			return dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if in.UserIds != nil {
			var ids []uuid.UUID
			for _, v := range in.UserIds {
				ids = append(ids, uuidx.ParseUUIDString(v))
			}
			err = tx.Role.Update().Where(role.ID(in.RoleId)).AddUserIDs(ids...).Exec(l.ctx)
			if err != nil {
				return err
			}

		} else {
			return errorx.NewInvalidArgumentError("userIds参数错误")
		}
		return nil

	})
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	return &core.BaseResp{Msg: "授权成功"}, nil
}
