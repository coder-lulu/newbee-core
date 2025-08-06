package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantLogic) UpdateTenant(req *types.TenantInfo) (resp *types.BaseMsgResp, err error) {
	// 转换TenantSettings类型
	var settings []*core.TenantSettings
	for _, setting := range req.Settings {
		if setting.Key != nil && setting.Value != nil {
			settings = append(settings, &core.TenantSettings{
				Key:   *setting.Key,
				Value: *setting.Value,
			})
		}
	}

	// 调用RPC服务更新租户
	result, err := l.svcCtx.CoreRpc.UpdateTenant(l.ctx, &core.TenantInfo{
		Id:            req.Id,
		Status:        req.Status,
		Name:          req.Name,
		Code:          req.Code,
		Domain:        req.Domain,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		Description:   req.Description,
		ExpireTime:    req.ExpireTime,
		MaxUsers:      req.MaxUsers,
		Settings:      settings,
	})

	if err != nil {
		return nil, err
	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, result.Msg)}, nil
}
