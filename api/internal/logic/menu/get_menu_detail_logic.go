package menu

import (
	"context"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMenuDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuDetailLogic {
	return &GetMenuDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *GetMenuDetailLogic) GetMenuDetail(req *types.IDReq) (resp *types.MenuPlainInfoResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetMenu(l.ctx, &core.IDReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	menuData := &types.MenuPlainInfo{
		Id:                 data.Id,
		MenuType:           data.MenuType,
		Level:              data.Level,
		Path:               data.Path,
		Name:               data.Name,
		Redirect:           data.Redirect,
		Component:          data.Component,
		Sort:               data.Sort,
		ParentId:           data.ParentId,
		Title:              data.Meta.Title,
		Icon:               data.Meta.Icon,
		HideMenu:           data.Meta.HideMenu,
		HideBreadcrumb:     data.Meta.HideBreadcrumb,
		IgnoreKeepAlive:    data.Meta.IgnoreKeepAlive,
		HideTab:            data.Meta.HideTab,
		FrameSrc:           data.Meta.FrameSrc,
		CarryParam:         data.Meta.CarryParam,
		HideChildrenInMenu: data.Meta.HideChildrenInMenu,
		Affix:              data.Meta.Affix,
		DynamicLevel:       data.Meta.DynamicLevel,
		RealPath:           data.Meta.RealPath,
		Disabled:           data.Disabled,
		ServiceName:        data.ServiceName,
		Permission:         data.Permission,
	}

	return &types.MenuPlainInfoResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  l.svcCtx.Trans.Trans(l.ctx, i18n.Success),
		},
		Data: *menuData,
	}, nil
}
