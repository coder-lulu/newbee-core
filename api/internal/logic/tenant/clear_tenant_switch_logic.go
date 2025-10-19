package tenant

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type ClearTenantSwitchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearTenantSwitchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearTenantSwitchLogic {
	return &ClearTenantSwitchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearTenantSwitchLogic) ClearTenantSwitch() (resp *types.BaseMsgResp, err error) {
	// 1. 获取当前用户信息
	userIdStr := l.svcCtx.ContextManager.GetUserID(l.ctx)
	if userIdStr == "" {
		return nil, errorx.NewApiError(http.StatusUnauthorized, "User not authenticated")
	}

	// 2. 清除Redis中的租户切换状态
	sessionKey := fmt.Sprintf("admin:tenant:%s", userIdStr)
	err = l.svcCtx.Redis.Del(l.ctx, sessionKey).Err()
	if err != nil {
		logx.Errorw("Failed to clear tenant switch from Redis", logx.Field("sessionKey", sessionKey), logx.Field("error", err))
		return nil, errorx.NewInternalError("Failed to clear tenant switch")
	}

	logx.Infow("Tenant switch cleared successfully", logx.Field("userId", userIdStr))

	return &types.BaseMsgResp{
		Code: 0,
		Msg:  "Tenant switch cleared successfully",
	}, nil
}
