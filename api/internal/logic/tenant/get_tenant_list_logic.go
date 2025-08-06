package tenant

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantListLogic) GetTenantList(req *types.TenantListReq) (resp *types.TenantListResp, err error) {
	// 调用RPC服务获取租户列表
	data, err := l.svcCtx.CoreRpc.GetTenantList(l.ctx, &core.TenantListReq{
		Page:          req.Page,
		PageSize:      req.PageSize,
		Name:          req.Name,
		Code:          req.Code,
		Domain:        req.Domain,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		ContactEmail:  req.ContactEmail,
		Description:   req.Description,
		ExpireTime:    req.ExpireTime,
		MaxUsers:      req.MaxUsers,
	})

	if err != nil {
		return nil, err
	}

	// 转换返回数据
	var tenantList []types.TenantInfo
	for _, v := range data.Data {
		// 转换Settings类型
		var settings []types.TenantSettings
		for _, setting := range v.Settings {
			settings = append(settings, types.TenantSettings{
				Key:   &setting.Key,
				Value: &setting.Value,
			})
		}

		tenantList = append(tenantList, types.TenantInfo{
			BaseIDInfo: types.BaseIDInfo{
				Id:        v.Id,
				CreatedAt: v.CreatedAt,
				UpdatedAt: v.UpdatedAt,
			},
			Status:        v.Status,
			Name:          v.Name,
			Code:          v.Code,
			Domain:        v.Domain,
			ContactPerson: v.ContactPerson,
			ContactPhone:  v.ContactPhone,
			ContactEmail:  v.ContactEmail,
			Description:   v.Description,
			ExpireTime:    v.ExpireTime,
			MaxUsers:      v.MaxUsers,
			Settings:      settings,
		})
	}

	return &types.TenantListResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "success",
		},
		Data: types.TenantListInfo{
			BaseListInfo: types.BaseListInfo{
				Total: data.Total,
			},
			Data: tenantList,
		},
	}, nil
}
