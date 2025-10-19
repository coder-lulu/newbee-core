package auditlog

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent/auditlog"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type CreateAuditLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAuditLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAuditLogLogic {
	return &CreateAuditLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAuditLogLogic) CreateAuditLog(in *core.AuditLogInfo) (*core.BaseUUIDResp, error) {
	// 🔍 调试日志：检查incoming context
	var reqTenantID string
	if in.TenantId != nil {
		reqTenantID = *in.TenantId
	}
	logx.Infow("CreateAuditLog - Incoming Context",
		logx.Field("tenant_id_from_request", reqTenantID))

	tenantIDStr, ok := l.ctx.Value(keys.TenantIDKey).(string)
	logx.Infow("CreateAuditLog - From Context Value",
		logx.Field("ok", ok),
		logx.Field("tenant_id", tenantIDStr))

	if !ok || tenantIDStr == "" {
		md, mdOK := metadata.FromIncomingContext(l.ctx)
		logx.Infow("CreateAuditLog - From Metadata",
			logx.Field("md_ok", mdOK),
			logx.Field("md_keys", func() []string {
				if mdOK {
					keys := make([]string, 0, len(md))
					for k := range md {
						keys = append(keys, k)
					}
					return keys
				}
				return nil
			}()))

		if mdOK {
			vals := md.Get(keys.TenantIDKey.String())
			logx.Infow("CreateAuditLog - Metadata tenant_id lookup",
				logx.Field("key", keys.TenantIDKey.String()),
				logx.Field("values", vals),
				logx.Field("len", len(vals)))

			if len(vals) > 0 {
				tenantIDStr = vals[0]
			}
		}
	}

	if tenantIDStr == "" {
		logx.Errorw("CreateAuditLog - tenant.missingContext",
			logx.Field("context_value", tenantIDStr),
			logx.Field("metadata_checked", "yes"))
		return nil, errorx.NewInvalidArgumentError("tenant.missingContext")
	}

	logx.Infow("✅ CreateAuditLog - tenant_id resolved",
		logx.Field("tenant_id", tenantIDStr))

	if in.TenantId != nil {
		if *in.TenantId != tenantIDStr {
			return nil, errorx.NewInvalidArgumentError("tenant.mismatch")
		}
	} else {
		in.TenantId = &tenantIDStr
	}

	logx.Infow("Creating audit log",
		logx.Field("tenantId", tenantIDStr),
		logx.Field("userId", in.UserId),
		logx.Field("path", in.RequestPath),
		logx.Field("method", in.RequestMethod))

	query := l.svcCtx.DB.AuditLog.Create().
		SetNotNilTenantID(pointy.GetPointer(tenantIDStr)).
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

	// Save using system context to bypass tenant check
	query = query.SetTenantID(tenantIDStr)

	result, err := query.Save(l.ctx)

	if err != nil {
		logx.Errorw("Failed to save audit log",
			logx.Field("error", err.Error()),
			logx.Field("tenantId", in.TenantId),
			logx.Field("userId", in.UserId))
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	logx.Infow("Audit log saved successfully",
		logx.Field("id", result.ID.String()),
		logx.Field("tenantId", in.TenantId),
		logx.Field("userId", in.UserId))

	return &core.BaseUUIDResp{Id: result.ID.String(), Msg: i18n.CreateSuccess}, nil
}
