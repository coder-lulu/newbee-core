package auditlog

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-core/rpc/ent/auditlog"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-common/v2/utils/uuidx"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAuditLogLogic {
	return &UpdateAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAuditLogLogic) UpdateAuditLog(in *core.AuditLogInfo) (*core.BaseResp, error) {
	query := l.svcCtx.DB.AuditLog.UpdateOneID(uuidx.ParseUUIDString(*in.Id)).
		SetNotNilTenantID(in.TenantId).
		SetNotNilUserID(in.UserId).
		SetNotNilUserName(in.UserName).
		SetNotNilResourceType(in.ResourceType).
		SetNotNilResourceID(in.ResourceId).
		SetNotNilRequestMethod(in.RequestMethod).
		SetNotNilRequestPath(in.RequestPath).
		SetNotNilRequestData(in.RequestData).
		SetNotNilResponseData(in.ResponseData).
		SetNotNilIPAddress(in.IpAddress).
		SetNotNilUserAgent(in.UserAgent).
		SetNotNilDurationMs(in.DurationMs).
		SetNotNilErrorMessage(in.ErrorMessage)

	// Handle OperationType conversion
	if in.OperationType != nil {
		query.SetOperationType(auditlog.OperationType(*in.OperationType))
	}

	// Handle Metadata conversion
	if in.Metadata != nil && *in.Metadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(*in.Metadata), &metadata); err == nil {
			query.SetMetadata(metadata)
		}
	}

	if in.Status != nil {
		query.SetNotNilStatus(pointy.GetPointer(uint8(*in.Status)))
	}
	if in.ResponseStatus != nil {
		query.SetNotNilResponseStatus(pointy.GetPointer(int(*in.ResponseStatus)))
	}

	err := query.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
