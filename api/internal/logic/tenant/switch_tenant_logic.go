package tenant

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type SwitchTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSwitchTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SwitchTenantLogic {
	return &SwitchTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SwitchTenantLogic) SwitchTenant(req *types.TenantSwitchReq) (resp *types.BaseMsgResp, err error) {
	// 1. 获取当前用户信息
	userIdStr := l.svcCtx.ContextManager.GetUserID(l.ctx)
	if userIdStr == "" {
		return nil, errorx.NewApiError(http.StatusUnauthorized, "User not authenticated")
	}

	// 2. 验证超级管理员权限
	isSuperAdmin, err := l.checkSuperAdminPermission(userIdStr)
	if err != nil {
		return nil, err
	}
	if !isSuperAdmin {
		return nil, errorx.NewApiError(http.StatusForbidden, "Only super admin can switch tenant")
	}

	// 3. 验证目标租户ID格式
	targetTenantId, err := strconv.ParseUint(req.TenantId, 10, 64)
	if err != nil {
		return nil, errorx.NewApiError(http.StatusBadRequest, "Invalid tenant ID format")
	}

	// 4. 验证目标租户是否存在且有效
	tenantInfo, err := l.svcCtx.CoreRpc.GetTenantById(l.ctx, &core.IDReq{
		Id: targetTenantId,
	})
	if err != nil {
		logx.Errorw("Failed to get tenant info", logx.Field("tenantId", req.TenantId), logx.Field("error", err))
		return nil, errorx.NewApiError(http.StatusBadRequest, "Target tenant not found")
	}

	if tenantInfo.Status == nil || *tenantInfo.Status != 1 { // 假设1为启用状态
		return nil, errorx.NewApiError(http.StatusBadRequest, "Target tenant is disabled")
	}

	// 5. 存储激活租户到Redis
	sessionKey := fmt.Sprintf("admin:tenant:%s", userIdStr)
	err = l.svcCtx.Redis.Set(l.ctx, sessionKey, req.TenantId, time.Hour*24).Err() // 24小时过期
	if err != nil {
		logx.Errorw("Failed to save tenant switch to Redis", logx.Field("sessionKey", sessionKey), logx.Field("error", err))
		return nil, errorx.NewInternalError("Failed to save tenant switch")
	}

	// 6. 记录审计日志
	l.logTenantSwitch(userIdStr, req.TenantId)

	logx.Infow("Tenant switched successfully",
		logx.Field("userId", userIdStr),
		logx.Field("targetTenantId", req.TenantId),
		logx.Field("tenantName", tenantInfo.Name))

	return &types.BaseMsgResp{
		Code: 0,
		Msg:  "Tenant switched successfully",
	}, nil
}

// checkSuperAdminPermission 检查用户是否为超级管理员
func (l *SwitchTenantLogic) checkSuperAdminPermission(userId string) (bool, error) {
	// 通过RPC获取用户信息
	userInfo, err := l.svcCtx.CoreRpc.GetUserById(l.ctx, &core.UUIDReq{
		Id: userId,
	})
	if err != nil {
		logx.Errorw("Failed to get user info", logx.Field("userId", userId), logx.Field("error", err))
		return false, errorx.NewInternalError("Failed to get user information")
	}

	// 检查角色码中是否包含superadmin
	for _, roleCode := range userInfo.RoleCodes {
		if strings.TrimSpace(roleCode) == "superadmin" {
			return true, nil
		}
	}

	return false, nil
}

// logTenantSwitch 记录租户切换的审计日志
func (l *SwitchTenantLogic) logTenantSwitch(userId, targetTenantId string) {
	// TODO: 实现审计日志记录
	logx.Infow("Tenant switch audit log",
		logx.Field("userId", userId),
		logx.Field("targetTenantId", targetTenantId),
		logx.Field("action", "TENANT_SWITCH"),
		logx.Field("timestamp", time.Now().Unix()))
}
