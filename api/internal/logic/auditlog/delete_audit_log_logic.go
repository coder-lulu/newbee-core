package auditlog

import (
	"context"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAuditLogLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAuditLogLogic {
	return &DeleteAuditLogLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAuditLogLogic) DeleteAuditLog(req *types.UUIDsReq) (resp *types.BaseMsgResp, err error) {
	_, err = l.svcCtx.CoreRpc.DeleteAuditLog(l.ctx, &core.UUIDsReq{
		Ids: req.Ids,
	})
	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{
		Code: 0,
		Msg:  l.svcCtx.Trans.Trans(l.ctx, i18n.DeleteSuccess),
	}, nil
}
