package auditlog

import (
	"context"
	"encoding/json"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuditLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuditLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogListLogic {
	return &GetAuditLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuditLogListLogic) GetAuditLogList(req *types.AuditLogListReq) (resp *types.AuditLogListResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetAuditLogList(l.ctx, &core.AuditLogListReq{
		Page:          uint64(req.Page),
		PageSize:      uint64(req.PageSize),
		UserId:        req.UserId,
		UserName:      req.UserName,
		OperationType: req.OperationType,
		ResourceType:  req.ResourceType,
		ResourceId:    req.ResourceId,
		RequestMethod: req.RequestMethod,
		RequestPath:   req.RequestPath,
		IpAddress:     req.IpAddress,
		ResponseStatus: func() *int64 {
			if req.ResponseStatus != nil {
				val := int64(*req.ResponseStatus)
				return &val
			}
			return nil
		}(),
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		MinDuration: req.MinDuration,
		MaxDuration: req.MaxDuration,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.AuditLogListResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  l.svcCtx.Trans.Trans(l.ctx, i18n.Success),
		},
		Data: types.AuditLogListInfo{
			BaseListInfo: types.BaseListInfo{
				Total: data.Total,
			},
			Data: []types.AuditLogInfo{},
		},
	}

	for _, v := range data.Data {
		// Convert metadata string back to map[string]interface{}
		var metadata map[string]interface{}
		if v.Metadata != nil && *v.Metadata != "" {
			if err := json.Unmarshal([]byte(*v.Metadata), &metadata); err != nil {
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		resp.Data.Data = append(resp.Data.Data, types.AuditLogInfo{
			BaseUUIDInfo: types.BaseUUIDInfo{
				Id:        v.Id,
				CreatedAt: v.CreatedAt,
				UpdatedAt: v.UpdatedAt,
			},
			Status:        v.Status,
			TenantId:      v.TenantId,
			UserId:        v.UserId,
			UserName:      v.UserName,
			OperationType: v.OperationType,
			ResourceType:  v.ResourceType,
			ResourceId:    v.ResourceId,
			RequestMethod: v.RequestMethod,
			RequestPath:   v.RequestPath,
			RequestData:   v.RequestData,
			ResponseStatus: func() *int32 {
				if v.ResponseStatus != nil {
					val := int32(*v.ResponseStatus)
					return &val
				}
				return nil
			}(),
			ResponseData: v.ResponseData,
			IpAddress:    v.IpAddress,
			UserAgent:    v.UserAgent,
			DurationMs:   v.DurationMs,
			ErrorMessage: v.ErrorMessage,
			Metadata:     metadata,
		})
	}

	return resp, nil
}
