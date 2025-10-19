package tenant

import (
	"context"
	"net/http"
	"strconv"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentTenantLogic {
	return &GetCurrentTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentTenantLogic) GetCurrentTenant() (resp *types.CurrentTenantResp, err error) {
	// 1. 获取当前用户信息
	userId := l.svcCtx.ContextManager.GetUserID(l.ctx)
	if userId == "" {
		return nil, errorx.NewApiError(http.StatusUnauthorized, "User not authenticated")
	}

	// 2. 从context获取当前/原始租户信息
	activeTenantIdStr := l.svcCtx.ContextManager.GetTenantID(l.ctx)
	if activeTenantIdStr == "" {
		return nil, errorx.NewInternalError("Tenant ID not found in context")
	}
	originalTenantIdStr := l.svcCtx.ContextManager.GetOriginalTenantID(l.ctx)
	if originalTenantIdStr == "" {
		originalTenantIdStr = activeTenantIdStr
	}

	// 3. 激活租户ID已在上面获取，这里不需要重复获取

	// 4. 判断是否已切换
	isSwitched := originalTenantIdStr != activeTenantIdStr

	// 5. 获取激活租户的详细信息
	var activeTenantInfo *types.TenantInfo
	if activeTenantIdStr != "" {
		activeTenantIdUint, parseErr := strconv.ParseUint(activeTenantIdStr, 10, 64)
		if parseErr == nil {
			tenantDetail, rpcErr := l.svcCtx.CoreRpc.GetTenantById(l.ctx, &core.IDReq{
				Id: activeTenantIdUint,
			})
			if rpcErr == nil {
				activeTenantInfo = &types.TenantInfo{
					BaseIDInfo: types.BaseIDInfo{
						Id:        tenantDetail.Id,
						CreatedAt: tenantDetail.CreatedAt,
						UpdatedAt: tenantDetail.UpdatedAt,
					},
					Status:      tenantDetail.Status,
					Name:        tenantDetail.Name,
					Code:        tenantDetail.Code,
					Description: tenantDetail.Description,
					ExpiredAt:   tenantDetail.ExpiredAt,
					Config:      tenantDetail.Config,
					CreatedBy:   tenantDetail.CreatedBy,
				}
			}
		}
	}

	return &types.CurrentTenantResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "success",
		},
		OriginalTenantId: originalTenantIdStr,
		ActiveTenantId:   activeTenantIdStr,
		ActiveTenantInfo: activeTenantInfo,
		IsSwitched:       isSwitched,
	}, nil
}
