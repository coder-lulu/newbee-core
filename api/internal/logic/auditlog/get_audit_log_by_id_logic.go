package auditlog

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuditLogByIdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuditLogByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogByIdLogic {
	return &GetAuditLogByIdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuditLogByIdLogic) GetAuditLogById(req *types.AuditLogReq) (resp *types.AuditLogResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetAuditLogById(l.ctx, &core.UUIDReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	// Convert metadata string back to map[string]interface{}
	var metadata map[string]interface{}
	if data.Metadata != nil && *data.Metadata != "" {
		if err := json.Unmarshal([]byte(*data.Metadata), &metadata); err != nil {
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	resp = &types.AuditLogResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  l.svcCtx.Trans.Trans(l.ctx, i18n.Success),
		},
		Data: types.AuditLogInfo{
			BaseUUIDInfo: types.BaseUUIDInfo{
				Id:        data.Id,
				CreatedAt: data.CreatedAt,
				UpdatedAt: data.UpdatedAt,
			},
			Status:        data.Status,
			TenantId:      data.TenantId,
			UserId:        data.UserId,
			UserName:      data.UserName,
			OperationType: data.OperationType,
			ResourceType:  data.ResourceType,
			ResourceId:    data.ResourceId,
			RequestMethod: data.RequestMethod,
			RequestPath:   data.RequestPath,
			RequestData:   data.RequestData,
			ResponseStatus: func() *int32 {
				if data.ResponseStatus != nil {
					val := int32(*data.ResponseStatus)
					return &val
				}
				return nil
			}(),
			ResponseData: data.ResponseData,
			IpAddress:    data.IpAddress,
			UserAgent:    data.UserAgent,
			DurationMs:   data.DurationMs,
			ErrorMessage: data.ErrorMessage,
			Metadata:     metadata,
		},
	}

	return resp, nil
}
