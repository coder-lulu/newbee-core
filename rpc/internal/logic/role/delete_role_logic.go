package role

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/config"

	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"

	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/v2/i18n"
)

type DeleteRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRoleLogic {
	return &DeleteRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteRoleLogic) DeleteRole(in *core.IDsReq) (*core.BaseResp, error) {
	count, err := l.svcCtx.DB.Role.Query().Where(role.HasUsers(), role.IDIn(in.Ids...)).Count(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if count != 0 {
		return nil, errorx.NewInvalidArgumentError("role.userExists")
	}

	_, err = l.svcCtx.DB.Role.Delete().Where(role.IDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = redisfunc.RemoveAllKeyByPrefix(l.ctx, fmt.Sprintf("%sROLE", config.RedisDataPermissionPrefix), l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &core.BaseResp{Msg: i18n.DeleteSuccess}, nil
}
