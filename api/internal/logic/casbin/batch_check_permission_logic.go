package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCheckPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCheckPermissionLogic {
	return &BatchCheckPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchCheckPermissionLogic) BatchCheckPermission(req *types.BatchPermissionCheckReq) (resp *types.BatchPermissionCheckResp, err error) {
	// 转换API类型为RPC类型
	rpcRequests := make([]*core.PermissionCheckReq, len(req.Requests))
	for i, apiReq := range req.Requests {
		rpcRequests[i] = &core.PermissionCheckReq{
			ServiceName: apiReq.ServiceName,
			Subject:     apiReq.Subject,
			Object:      apiReq.Object,
			Action:      apiReq.Action,
			Context:     apiReq.Context,
			EnableCache: apiReq.EnableCache,
			AuditLog:    apiReq.AuditLog,
		}
	}

	rpcReq := &core.BatchPermissionCheckReq{
		Requests: rpcRequests,
		FailFast: req.FailFast,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.BatchCheckPermission(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to batch check permission via RPC: %v", err)
		return &types.BatchPermissionCheckResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "批量权限检查失败: " + err.Error(),
			},
			Data: types.BatchPermissionCheckResult{
				Responses:    []types.PermissionCheckResult{},
				SuccessCount: 0,
				FailedCount:  int32(len(req.Requests)),
			},
		}, nil
	}

	// 转换RPC响应为API类型
	apiResponses := make([]types.PermissionCheckResult, len(rpcResp.Responses))
	for i, rpcResult := range rpcResp.Responses {
		apiResponses[i] = types.PermissionCheckResult{
			Allowed:         rpcResult.Allowed,
			Reason:          rpcResult.Reason,
			AppliedRules:    rpcResult.AppliedRules,
			DataFilters:     rpcResult.DataFilters,
			FieldMasks:      rpcResult.FieldMasks,
			CheckDurationMs: rpcResult.CheckDurationMs,
			FromCache:       rpcResult.FromCache,
		}
	}

	return &types.BatchPermissionCheckResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "批量权限检查完成",
		},
		Data: types.BatchPermissionCheckResult{
			Responses:    apiResponses,
			SuccessCount: rpcResp.SuccessCount,
			FailedCount:  rpcResp.FailedCount,
		},
	}, nil
}
