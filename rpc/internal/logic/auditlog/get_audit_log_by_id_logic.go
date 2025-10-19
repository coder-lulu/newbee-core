package auditlog

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetAuditLogByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuditLogByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogByIdLogic {
	return &GetAuditLogByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAuditLogByIdLogic) GetAuditLogById(in *core.UUIDReq) (*core.AuditLogInfo, error) {
	tenantIDStr, ok := l.ctx.Value(keys.TenantIDKey).(string)
	if !ok || tenantIDStr == "" {
		if md, mdOK := metadata.FromIncomingContext(l.ctx); mdOK {
			if vals := md.Get(keys.TenantIDKey.String()); len(vals) > 0 {
				tenantIDStr = vals[0]
			}
		}
	}

	if tenantIDStr == "" {
		return nil, errorx.NewInvalidArgumentError("tenant.missingContext")
	}

	result, err := l.svcCtx.DB.AuditLog.Get(l.ctx, uuidx.ParseUUIDString(in.Id))
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if result.TenantID != tenantIDStr {
		return nil, errorx.NewInvalidArgumentError("tenant.mismatch")
	}

	userName := result.UserName
	if result.UserID != "" {
		if userUUID := uuidx.ParseUUIDString(result.UserID); userUUID.String() != "00000000-0000-0000-0000-000000000000" {
			tenantIDUint, err := strconv.ParseUint(tenantIDStr, 10, 64)
			if err != nil {
				return nil, errorx.NewInvalidArgumentError("tenant.invalidContext")
			}

			if userInfo, err := l.svcCtx.DB.User.Query().
				Where(user.IDEQ(userUUID), user.TenantIDEQ(tenantIDUint)).
				First(l.ctx); err == nil {
				if userInfo.Username != "" {
					userName = userInfo.Username
				} else if userInfo.Nickname != "" {
					userName = userInfo.Nickname
				} else if userInfo.Email != "" {
					userName = userInfo.Email
				} else {
					userName = "[用户名为空]"
				}

				l.Logger.Infow("Retrieved username for audit detail",
					logx.Field("userID", result.UserID),
					logx.Field("username", userName))
			} else {
				userName = "[用户不存在]"
				l.Logger.Errorw("User not found for audit detail",
					logx.Field("userID", result.UserID),
					logx.Field("error", err))
			}
		}
	}

	tenantIDCopy := result.TenantID

	return &core.AuditLogInfo{
		Id:             pointy.GetPointer(result.ID.String()),
		CreatedAt:      pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:      pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Status:         pointy.GetPointer(uint32(result.Status)),
		TenantId:       &tenantIDCopy,
		UserId:         &result.UserID,
		UserName:       &userName, // 使用查询得到的用户名
		OperationType:  pointy.GetPointer(string(result.OperationType)),
		ResourceType:   &result.ResourceType,
		ResourceId:     &result.ResourceID,
		RequestMethod:  &result.RequestMethod,
		RequestPath:    &result.RequestPath,
		RequestData:    &result.RequestData,
		ResponseStatus: pointy.GetPointer(int64(result.ResponseStatus)),
		ResponseData:   &result.ResponseData,
		IpAddress:      &result.IPAddress,
		UserAgent:      &result.UserAgent,
		DurationMs:     &result.DurationMs,
		ErrorMessage:   &result.ErrorMessage,
		Metadata: func() *string {
			if result.Metadata != nil {
				if jsonData, err := json.Marshal(result.Metadata); err == nil {
					return pointy.GetPointer(string(jsonData))
				}
			}
			return pointy.GetPointer("")
		}(),
	}, nil
}
