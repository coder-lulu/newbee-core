package casbin

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPermissionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPermissionLogic {
	return &CheckPermissionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckPermissionLogic) CheckPermission(req *types.PermissionCheckReq) (resp *types.PermissionCheckResp, err error) {
	// 将API类型转换为RPC类型
	rpcReq := &core.PermissionCheckReq{
		ServiceName: req.ServiceName,
		Subject:     req.Subject,
		Object:      req.Object,
		Action:      req.Action,
		Context:     req.Context,
		EnableCache: req.EnableCache,
		AuditLog:    req.AuditLog,
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.CoreRpc.CheckPermission(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("Failed to check permission via RPC: %v", err)
		return &types.PermissionCheckResp{
			BaseDataInfo: types.BaseDataInfo{
				Code: 1,
				Msg:  "权限检查失败: " + err.Error(),
			},
			Data: types.PermissionCheckResult{
				Allowed:         false,
				Reason:          "系统错误",
				AppliedRules:    []string{},
				DataFilters:     make(map[string]string),
				FieldMasks:      []string{},
				CheckDurationMs: 0,
				FromCache:       false,
			},
		}, nil
	}

	// 构建成功响应
	return &types.PermissionCheckResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "权限检查完成",
		},
		Data: types.PermissionCheckResult{
			Allowed:         rpcResp.Allowed,
			Reason:          rpcResp.Reason,
			AppliedRules:    rpcResp.AppliedRules,
			DataFilters:     rpcResp.DataFilters,
			FieldMasks:      rpcResp.FieldMasks,
			CheckDurationMs: rpcResp.CheckDurationMs,
			FromCache:       rpcResp.FromCache,
		},
	}, nil
}
