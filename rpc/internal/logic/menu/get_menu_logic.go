package menu

import (
	"context"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuLogic {
	return &GetMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuLogic) GetMenu(in *core.IDReq) (*core.MenuInfo, error) {
	menu, err := l.svcCtx.DB.Menu.Query().Where(menu.ID(in.Id)).First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	data := &core.MenuInfo{
		Id:          &menu.ID,
		CreatedAt:   pointy.GetPointer(menu.CreatedAt.UnixMilli()),
		UpdatedAt:   pointy.GetPointer(menu.UpdatedAt.UnixMilli()),
		MenuType:    &menu.MenuType,
		Level:       &menu.MenuLevel,
		ParentId:    &menu.ParentID,
		Path:        &menu.Path,
		Name:        &menu.Name,
		Redirect:    &menu.Redirect,
		Component:   &menu.Component,
		Disabled:    &menu.Disabled,
		Sort:        &menu.Sort,
		ServiceName: &menu.ServiceName,
		Permission:  &menu.Permission,
		Meta: &core.Meta{
			Title:              &menu.Title,
			Icon:               &menu.Icon,
			HideMenu:           &menu.HideMenu,
			HideBreadcrumb:     &menu.HideBreadcrumb,
			IgnoreKeepAlive:    &menu.IgnoreKeepAlive,
			HideTab:            &menu.HideTab,
			FrameSrc:           &menu.FrameSrc,
			CarryParam:         &menu.CarryParam,
			HideChildrenInMenu: &menu.HideChildrenInMenu,
			Affix:              &menu.Affix,
			DynamicLevel:       &menu.DynamicLevel,
			RealPath:           &menu.RealPath,
			Params:             &menu.Params,
		},
	}

	return data, nil
}
