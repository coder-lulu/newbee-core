package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"

	"entgo.io/ent/dialect/sql"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTenantLogic {
	return &DeleteTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTenantLogic) DeleteTenant(in *core.IDsReq) (*core.BaseResp, error) {
	if len(in.Ids) == 0 {
		return nil, errorx.NewInvalidArgumentError("tenant.idsRequired")
	}

	// 检查是否包含默认租户（ID为1的租户不能删除）
	for _, id := range in.Ids {
		if id == 1 {
			return nil, errorx.NewInvalidArgumentError("tenant.cannotDeleteDefault")
		}
	}

	// 使用系统上下文查询租户是否存在关联数据
	// 注意：由于租户隔离，这里使用系统上下文来检查所有数据
	systemCtx := hooks.NewSystemContext(l.ctx)

	for _, id := range in.Ids {
		// 检查是否有用户属于此租户
		userCount, err := l.svcCtx.DB.User.Query().
			Where(func(selector *sql.Selector) {
				selector.Where(sql.EQ(selector.C("tenant_id"), id))
			}).
			Count(systemCtx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if userCount > 0 {
			return nil, errorx.NewInvalidArgumentError("tenant.hasUsers")
		}

		// 检查是否有其他关联数据（角色、部门等）
		roleCount, err := l.svcCtx.DB.Role.Query().
			Where(func(selector *sql.Selector) {
				selector.Where(sql.EQ(selector.C("tenant_id"), id))
			}).
			Count(systemCtx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if roleCount > 0 {
			return nil, errorx.NewInvalidArgumentError("tenant.hasRoles")
		}
	}

	// 使用系统上下文删除租户
	_, err := l.svcCtx.DB.Tenant.Delete().
		Where(tenant.IDIn(in.Ids...)).
		Exec(hooks.NewSystemContext(l.ctx))

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	logx.Infow("Tenants deleted successfully",
		logx.Field("tenant_ids", in.Ids))

	return &core.BaseResp{
		Msg: i18n.DeleteSuccess,
	}, nil
}
