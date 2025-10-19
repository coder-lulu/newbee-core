package department

import (
	"context"
	"strconv"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/orm/ent/entctx/datapermctx"
	"github.com/coder-lulu/newbee-common/orm/ent/entctx/deptctx"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dbfunc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type InitDeptDataPermToRedisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInitDeptDataPermToRedisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitDeptDataPermToRedisLogic {
	return &InitDeptDataPermToRedisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *InitDeptDataPermToRedisLogic) InitDeptDataPermToRedis(in *core.Empty) (*core.BaseResp, error) {
	deptId, err := deptctx.GetDepartmentIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	subDept, err := dbfunc.GetSubDepartment(deptId, l.svcCtx.DB, l.Logger, l.ctx)
	if err != nil {
		return nil, err
	}

	if subDept == "" {
		subDept = strconv.Itoa(int(deptId))
	}

	err = l.svcCtx.Redis.Set(l.ctx, datapermctx.GetSubDeptDataPermRedisKey(deptId), subDept, redis.KeepTTL).Err()
	if err != nil {
		logx.Error("failed to set sub department data to redis for data permission", logx.Field("detail", err))
		return nil, errorx.NewInternalError(i18n.RedisError)
	}

	return &core.BaseResp{Msg: i18n.Success}, nil
}
