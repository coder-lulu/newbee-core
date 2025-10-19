package auditlog

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/ent/auditlog"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/utils/uuidx"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAuditLogLogic {
	return &DeleteAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteAuditLogLogic) DeleteAuditLog(in *core.UUIDsReq) (*core.BaseResp, error) {
	_, err := l.svcCtx.DB.AuditLog.Delete().Where(auditlog.IDIn(uuidx.ParseUUIDSlice(in.Ids)...)).Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.DeleteSuccess}, nil
}
