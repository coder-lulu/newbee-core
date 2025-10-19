package menu

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuListByRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuListByRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuListByRoleLogic {
	return &GetMenuListByRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMenuListByRoleLogic) GetMenuListByRole() (resp *types.MenuListResp, err error) {
	// 从context获取角色编码
	roleCodes := ""
	if val := l.ctx.Value(keys.RoleCodesKey); val != nil {
		if codes, ok := val.(string); ok {
			roleCodes = codes
		}
	}
	data, err := l.svcCtx.CoreRpc.GetMenuListByRole(l.ctx, &core.BaseMsg{Msg: roleCodes})
	if err != nil {
		return nil, err
	}
	resp = &types.MenuListResp{}
	resp.Data.Total = data.Total
	if data.Total == 0 {
		resp.Data.Data = []types.MenuInfo{}
		return resp, nil
	}
	for _, v := range data.Data {
		resp.Data.Data = append(resp.Data.Data, types.MenuInfo{
			BaseIDInfo: types.BaseIDInfo{
				Id: v.Id,
			},
			MenuType:    v.MenuType,
			Level:       v.Level,
			Path:        v.Path,
			Name:        v.Name,
			Redirect:    v.Redirect,
			Component:   v.Component,
			Sort:        v.Sort,
			ParentId:    v.ParentId,
			ServiceName: v.ServiceName,
			Permission:  v.Permission,
			Meta: types.Meta{
				Title:              pointy.GetPointer(l.svcCtx.Trans.Trans(l.ctx, *v.Meta.Title)),
				Icon:               v.Meta.Icon,
				HideMenu:           v.Meta.HideMenu,
				HideBreadcrumb:     v.Meta.HideBreadcrumb,
				IgnoreKeepAlive:    v.Meta.IgnoreKeepAlive,
				HideTab:            v.Meta.HideTab,
				FrameSrc:           v.Meta.FrameSrc,
				CarryParam:         v.Meta.CarryParam,
				HideChildrenInMenu: v.Meta.HideChildrenInMenu,
				Affix:              v.Meta.Affix,
				DynamicLevel:       v.Meta.DynamicLevel,
				RealPath:           v.Meta.RealPath,
				Params:             v.Meta.Params,
			},
		})
	}
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, i18n.Success)
	return resp, nil
}
