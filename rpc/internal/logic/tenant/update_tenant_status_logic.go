package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantStatusLogic {
	return &UpdateTenantStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTenantStatusLogic) UpdateTenantStatus(in *core.TenantStatusReq) (*core.BaseResp, error) {
	// 验证租户是否存在
	exist, err := l.svcCtx.DB.Tenant.Query().Where(tenant.IDEQ(in.Id)).Exist(hooks.NewSystemContext(l.ctx))
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if !exist {
		return nil, errorx.NewInvalidArgumentError("tenant.notFound")
	}

	// 不允许禁用默认租户
	if in.Id == 1 && in.Status == 2 {
		return nil, errorx.NewInvalidArgumentError("tenant.cannotDisableDefault")
	}

	// 使用系统上下文更新租户状态
	_, err = l.svcCtx.DB.Tenant.UpdateOneID(in.Id).
		SetStatus(uint8(in.Status)).
		Save(hooks.NewSystemContext(l.ctx))

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	logx.Infow("Tenant status updated successfully",
		logx.Field("tenant_id", in.Id),
		logx.Field("new_status", in.Status))

	return &core.BaseResp{
		Msg: i18n.UpdateSuccess,
	}, nil
}
