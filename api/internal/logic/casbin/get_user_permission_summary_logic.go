package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermissionSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserPermissionSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermissionSummaryLogic {
	return &GetUserPermissionSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserPermissionSummaryLogic) GetUserPermissionSummary(req *types.UserPermissionSummaryReq) (resp *types.UserPermissionSummaryResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.GetUserPermissionSummaryReq{
		UserId:           req.UserId,
		ServiceName:      req.ServiceName,
		IncludeInherited: req.IncludeInherited,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.GetUserPermissionSummary(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to get user permission summary via RPC: %v", err)
		return &types.UserPermissionSummaryResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "获取用户权限摘要失败: " + err.Error(),
			},
		}, nil
	}

	// 将RPC类型转换为API类型
	var permissions []types.PermissionSummary
	for _, rpcPerm := range rpcResp.Permissions {
		apiPerm := types.PermissionSummary{
			Resource: rpcPerm.Resource,
			Actions:  rpcPerm.Actions,
			Source:   rpcPerm.Source,
			RuleId:   rpcPerm.RuleId,
		}
		permissions = append(permissions, apiPerm)
	}

	apiData := types.UserPermissionSummary{
		UserId:      rpcResp.UserId,
		Permissions: permissions,
		TotalCount:  rpcResp.TotalCount,
	}

	// 构建成功响应
	return &types.UserPermissionSummaryResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "获取用户权限摘要成功",
		},
		Data: apiData,
	}, nil
}
