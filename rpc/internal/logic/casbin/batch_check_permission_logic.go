package casbin

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCheckPermissionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchCheckPermissionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCheckPermissionLogic {
	return &BatchCheckPermissionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchCheckPermissionLogic) BatchCheckPermission(in *core.BatchPermissionCheckReq) (*core.BatchPermissionCheckResp, error) {
	// 验证请求
	if len(in.Requests) == 0 {
		return nil, fmt.Errorf("requests cannot be empty")
	}

	// 批量处理权限检查
	var responses []*core.PermissionCheckResp
	var successCount, failedCount int32

	checkLogic := NewCheckPermissionLogic(l.ctx, l.svcCtx)

	for i, req := range in.Requests {
		resp, err := checkLogic.CheckPermission(req)
		if err != nil {
			failedCount++
			// 如果启用快速失败模式，遇到错误立即返回
			if in.FailFast != nil && *in.FailFast {
				return nil, fmt.Errorf("batch permission check failed at request %d: %v", i, err)
			}

			// 创建失败响应
			resp = &core.PermissionCheckResp{
				Allowed:         false,
				Reason:          fmt.Sprintf("error: %v", err),
				AppliedRules:    []string{},
				DataFilters:     make(map[string]string),
				FieldMasks:      []string{},
				CheckDurationMs: 0,
				FromCache:       false,
			}
		} else {
			successCount++
		}

		responses = append(responses, resp)
	}

	l.Logger.Infof("Batch permission check completed: %d total, %d success, %d failed",
		len(in.Requests), successCount, failedCount)

	return &core.BatchPermissionCheckResp{
		Responses:    responses,
		SuccessCount: successCount,
		FailedCount:  failedCount,
	}, nil
}
