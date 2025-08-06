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

type CreateTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateTenantLogic) CreateTenant(in *core.TenantInfo) (*core.BaseIDResp, error) {
	// 转换TenantSettings类型
	var settings []types.TenantSettings
	for _, setting := range in.Settings {
		settings = append(settings, types.TenantSettings{
			Key:   setting.Key,
			Value: setting.Value,
		})
	}

	query := l.svcCtx.DB.Tenant.Create().
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

	result, err := query.Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 自动初始化新租户的基础数据
	initLogic := NewInitTenantDataLogic(l.ctx, l.svcCtx)
	_, initErr := initLogic.InitTenantData(&core.IDReq{Id: result.ID})
	if initErr != nil {
		logx.Errorw("Failed to initialize tenant data",
			logx.Field("tenantId", result.ID),
			logx.Field("error", initErr))
		// 注意：这里不返回错误，因为租户已经创建成功，只是初始化失败
		// 可以通过手动调用初始化接口来补救
	}

	return &core.BaseIDResp{Id: result.ID, Msg: i18n.CreateSuccess}, nil
}
