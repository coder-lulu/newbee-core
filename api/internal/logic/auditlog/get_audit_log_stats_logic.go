package auditlog

import (
	"context"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuditLogStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAuditLogStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuditLogStatsLogic {
	return &GetAuditLogStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAuditLogStatsLogic) GetAuditLogStats(req *types.AuditLogStatsReq) (resp *types.AuditLogStatsResp, err error) {
	// Determine what to group by
	groupBy := "operation_type"
	if req.GroupBy != nil {
		groupBy = *req.GroupBy
	}

	// Create the RPC request - simplified approach
	data, err := l.svcCtx.CoreRpc.GetAuditLogStats(l.ctx, &core.AuditLogStatsReq{
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.AuditLogStatsResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  l.svcCtx.Trans.Trans(l.ctx, i18n.Success),
		},
		Data: []types.AuditLogStatsItem{},
	}

	// Convert stats based on groupBy parameter
	switch groupBy {
	case "operation_type":
		for _, v := range data.OperationStats {
			resp.Data = append(resp.Data, types.AuditLogStatsItem{
				Label:      v.OperationType,
				Count:      int64(v.Count),
				Percentage: v.Percentage,
			})
		}
	case "resource_type":
		for _, v := range data.ResourceStats {
			resp.Data = append(resp.Data, types.AuditLogStatsItem{
				Label:      v.ResourceType,
				Count:      int64(v.Count),
				Percentage: v.Percentage,
			})
		}
	case "duration":
		for _, v := range data.DurationStats {
			resp.Data = append(resp.Data, types.AuditLogStatsItem{
				Label:      v.RangeLabel,
				Count:      int64(v.Count),
				Percentage: v.Percentage,
			})
		}
	default:
		// Default to operation type stats
		for _, v := range data.OperationStats {
			resp.Data = append(resp.Data, types.AuditLogStatsItem{
				Label:      v.OperationType,
				Count:      int64(v.Count),
				Percentage: v.Percentage,
			})
		}
	}

	return resp, nil
}
