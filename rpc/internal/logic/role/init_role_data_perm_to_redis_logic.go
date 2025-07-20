package role

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/redis/go-redis/v9"
	"github.com/suyuan32/simple-admin-common/i18n"
	"github.com/suyuan32/simple-admin-common/orm/ent/entctx/datapermctx"
	"github.com/suyuan32/simple-admin-common/orm/ent/entctx/rolectx"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

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
	roleCodes, err := rolectx.GetRoleIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	var customDeptIds []string
	dataScope := uint8(math.MaxUint8)

	for _, v := range roleCodes {
		roleData, err := l.svcCtx.DB.Role.Query().Where(role.CodeEQ(v)).Only(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if roleData.DataScope < dataScope {
			dataScope = roleData.DataScope
		}

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

	if dataScope != uint8(math.MaxUint8) {
		err = l.svcCtx.Redis.Set(l.ctx, datapermctx.GetRoleScopeDataPermRedisKey(roleCodes), strconv.Itoa(int(dataScope)), redis.KeepTTL).Err()
		if err != nil {
			logx.Error("failed to set role scope data to redis for data permission", logx.Field("detail", err))
			return nil, errorx.NewInternalError(i18n.RedisError)
		}
	}

	return &core.BaseResp{Msg: i18n.Success}, nil
}
