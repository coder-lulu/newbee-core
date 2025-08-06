package tenant

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/tenant"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantListLogic) GetTenantList(in *core.TenantListReq) (*core.TenantListResp, error) {
	var predicates []predicate.Tenant
	if in.CreatedAt != nil {
		predicates = append(predicates, tenant.CreatedAtGTE(time.UnixMilli(*in.CreatedAt)))
	}
	if in.UpdatedAt != nil {
		predicates = append(predicates, tenant.UpdatedAtGTE(time.UnixMilli(*in.UpdatedAt)))
	}
	if in.Status != nil {
		predicates = append(predicates, tenant.StatusEQ(uint8(*in.Status)))
	}
	if in.DeletedAt != nil {
		predicates = append(predicates, tenant.DeletedAtGTE(time.UnixMilli(*in.DeletedAt)))
	}
	if in.Name != nil {
		predicates = append(predicates, tenant.NameContains(*in.Name))
	}
	if in.Code != nil {
		predicates = append(predicates, tenant.CodeContains(*in.Code))
	}
	if in.Domain != nil {
		predicates = append(predicates, tenant.DomainContains(*in.Domain))
	}
	if in.ContactPerson != nil {
		predicates = append(predicates, tenant.ContactPersonContains(*in.ContactPerson))
	}
	if in.ContactPhone != nil {
		predicates = append(predicates, tenant.ContactPhoneContains(*in.ContactPhone))
	}
	if in.ContactEmail != nil {
		predicates = append(predicates, tenant.ContactEmailContains(*in.ContactEmail))
	}
	if in.Description != nil {
		predicates = append(predicates, tenant.DescriptionContains(*in.Description))
	}
	if in.ExpireTime != nil {
		predicates = append(predicates, tenant.ExpireTimeGTE(time.UnixMilli(*in.ExpireTime)))
	}
	if in.MaxUsers != nil {
		predicates = append(predicates, tenant.MaxUsersEQ(*in.MaxUsers))
	}
	// Settings字段不支持直接搜索，跳过
	result, err := l.svcCtx.DB.Tenant.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.TenantListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		// 转换Settings类型
		var settings []*core.TenantSettings
		for _, setting := range v.Settings {
			settings = append(settings, &core.TenantSettings{
				Key:   setting.Key,
				Value: setting.Value,
			})
		}

		resp.Data = append(resp.Data, &core.TenantInfo{
			Id:            &v.ID,
			CreatedAt:     pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:     pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Status:        pointy.GetPointer(uint32(v.Status)),
			Name:          &v.Name,
			Code:          &v.Code,
			Domain:        &v.Domain,
			ContactPerson: &v.ContactPerson,
			ContactPhone:  &v.ContactPhone,
			ContactEmail:  &v.ContactEmail,
			Description:   &v.Description,
			ExpireTime:    pointy.GetUnixMilliPointer(v.ExpireTime.UnixMilli()),
			MaxUsers:      &v.MaxUsers,
			Settings:      settings,
		})
	}

	return resp, nil
}
