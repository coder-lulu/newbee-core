package role

import (
	"context"
	"strconv"
	"strings"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/datapermctx"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/rolectx"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/redis/go-redis/v9"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type InitRoleDataPermToRedisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitRoleDataPermToRedisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitRoleDataPermToRedisLogic {
	return &InitRoleDataPermToRedisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitRoleDataPermToRedisLogic) InitRoleDataPermToRedis(in *core.Empty) (*core.BaseResp, error) {
	// Try to get role codes from context first
	var roleCodes []string
	var err error

	if roleCodesStr, ok := l.ctx.Value("roleCodes").(string); ok && roleCodesStr != "" {
		roleCodes = strings.Split(roleCodesStr, ",")
		for i, role := range roleCodes {
			roleCodes[i] = strings.TrimSpace(role)
		}
		logx.Infow("Using roleCodes from context", logx.Field("roleCodes", roleCodes))
	} else {
		// Fallback: try to get from roleId context (this may contain role codes for compatibility)
		roleCodes, err = rolectx.GetRoleIDFromCtx(l.ctx)
		if err != nil {
			logx.Errorw("Failed to get role information from context", logx.Field("error", err), logx.Field("context", l.ctx))
			return nil, err
		}
		logx.Infow("Using roleIds from context as roleCodes", logx.Field("roleCodes", roleCodes))
	}

	// Ensure tenant context is available for database queries
	// This should have tenant info from API call, but ensure it's properly set
	queryCtx := l.ctx
	if tenantId := tenantctx.GetTenantIDFromCtx(l.ctx); tenantId > 0 {
		// Context already has tenant info from API call
		// 使用统一的ContextManager设置租户ID
		cm := keys.NewContextManager()
		tenantIdStr := strconv.FormatUint(tenantId, 10)
		queryCtx = cm.SetTenantID(l.ctx, tenantIdStr)
	}

	// 🔥 Phase 3: data_scope field removed from sys_roles table
	// Data permission is now managed via sys_casbin_rules (ptype='d')
	// This legacy Redis caching logic is deprecated and should be rewritten
	// TODO: Reimplement this method to query data permissions from sys_casbin_rules
	//       instead of sys_roles.data_scope

	var customDeptIds []string

	for _, v := range roleCodes {
		roleData, err := l.svcCtx.DB.Role.Query().Where(role.CodeEQ(v)).Only(queryCtx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		// Still process custom_dept_ids as this field remains in sys_roles
		for _, id := range roleData.CustomDeptIds {
			customDeptIds = append(customDeptIds, strconv.Itoa(int(id)))
		}
	}

	customDeptIds = slice.Unique(customDeptIds)
	slice.Sort(customDeptIds)

	if len(customDeptIds) > 0 {
		err = l.svcCtx.Redis.Set(l.ctx, datapermctx.GetRoleCustomDeptDataPermRedisKey(roleCodes), strings.Join(customDeptIds, ","), redis.KeepTTL).Err()
		if err != nil {
			logx.Error("failed to set role custom department data to redis for data permission", logx.Field("detail", err))
			return nil, errorx.NewInternalError(i18n.RedisError)
		}
	}

	// 🔥 Phase 3: dataScope storage removed
	// Data scope is now queried from sys_casbin_rules at runtime via UnifiedDataPermPlugin
	logx.Info("✅ Custom department IDs cached to Redis. Data scope management moved to Casbin.")

	return &core.BaseResp{Msg: i18n.Success}, nil
}
