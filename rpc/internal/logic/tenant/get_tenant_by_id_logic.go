package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantByIdLogic {
	return &GetTenantByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantByIdLogic) GetTenantById(in *core.IDReq) (*core.TenantInfo, error) {
	result, err := l.svcCtx.DB.Tenant.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 转换Settings类型
	var settings []*core.TenantSettings
	for _, setting := range result.Settings {
		settings = append(settings, &core.TenantSettings{
			Key:   setting.Key,
			Value: setting.Value,
		})
	}

	return &core.TenantInfo{
		Id:            &result.ID,
		CreatedAt:     pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:     pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Status:        pointy.GetPointer(uint32(result.Status)),
		Name:          &result.Name,
		Code:          &result.Code,
		Domain:        &result.Domain,
		ContactPerson: &result.ContactPerson,
		ContactPhone:  &result.ContactPhone,
		ContactEmail:  &result.ContactEmail,
		Description:   &result.Description,
		ExpireTime:    pointy.GetUnixMilliPointer(result.ExpireTime.UnixMilli()),
		MaxUsers:      &result.MaxUsers,
		Settings:      settings,
	}, nil
}
