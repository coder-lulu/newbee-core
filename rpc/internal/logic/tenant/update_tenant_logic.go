package tenant

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/rpc/ent/schema/types"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTenantLogic) UpdateTenant(in *core.TenantInfo) (*core.BaseResp, error) {
	// 转换Settings类型
	var settings []types.TenantSettings
	for _, setting := range in.Settings {
		settings = append(settings, types.TenantSettings{
			Key:   setting.Key,
			Value: setting.Value,
		})
	}

	query := l.svcCtx.DB.Tenant.UpdateOneID(*in.Id).
		SetNotNilName(in.Name).
		SetNotNilCode(in.Code).
		SetNotNilDomain(in.Domain).
		SetNotNilContactPerson(in.ContactPerson).
		SetNotNilContactPhone(in.ContactPhone).
		SetNotNilContactEmail(in.ContactEmail).
		SetNotNilDescription(in.Description).
		SetNotNilExpireTime(func() *time.Time {
			if in.ExpireTime != nil {
				t := time.UnixMilli(*in.ExpireTime)
				return &t
			}
			return nil
		}()).
		SetNotNilMaxUsers(in.MaxUsers).
		SetNotNilSettings(settings)

	if in.Status != nil {
		status := uint8(*in.Status)
		query.SetNotNilStatus(&status)
	}

	err := query.Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
