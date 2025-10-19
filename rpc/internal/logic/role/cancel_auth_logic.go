package role

import (
	"context"

	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"
	uuid "github.com/gofrs/uuid/v5"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelAuthLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelAuthLogic {
	return &CancelAuthLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelAuthLogic) CancelAuth(in *core.RoleAuthReq) (*core.BaseResp, error) {
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
			tx.Role.UpdateOneID(in.RoleId).RemoveUserIDs(ids...).Exec(l.ctx)
			if err != nil {
				return dberrorhandler.DefaultEntError(l.Logger, err, in)
			}
			// err = tx.Role.Update().Where(role.ID(in.RoleId)).Where(role.IDIn()).ClearUsers().Exec(l.ctx)
			// if err != nil {
			// 	return err
			// }

		} else {
			return errorx.NewInvalidArgumentError("userIds参数错误")
		}
		return nil

	})
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	return &core.BaseResp{Msg: "取消成功"}, nil
}
