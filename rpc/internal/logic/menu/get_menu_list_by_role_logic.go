package menu

import (
	"context"
	"strings"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuListByRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuListByRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuListByRoleLogic {
	return &GetMenuListByRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuListByRoleLogic) GetMenuListByRole(in *core.BaseMsg) (*core.MenuInfoList, error) {
	roles, err := l.svcCtx.DB.Role.Query().Where(role.CodeIn(strings.Split(in.Msg, ",")...), role.StatusEQ(1)).WithMenus(func(query *ent.MenuQuery) {
		query.Order(ent.Asc(menu.FieldSort))
		query.Where(menu.Disabled(false))
	}).All(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.MenuInfoList{}

	existMap := map[uint64]struct{}{}
	for _, r := range roles {
		for _, m := range r.Edges.Menus {
			if _, ok := existMap[m.ID]; !ok {
				resp.Data = append(resp.Data, &core.MenuInfo{
					Id:          &m.ID,
					CreatedAt:   pointy.GetPointer(m.CreatedAt.UnixMilli()),
					UpdatedAt:   pointy.GetPointer(m.UpdatedAt.UnixMilli()),
					MenuType:    &m.MenuType,
					Level:       &m.MenuLevel,
					ParentId:    &m.ParentID,
					Path:        &m.Path,
					Name:        &m.Name,
					Redirect:    &m.Redirect,
					Component:   &m.Component,
					Sort:        &m.Sort,
					ServiceName: &m.ServiceName,
					Permission:  &m.Permission,
					Meta: &core.Meta{
						Title:              &m.Title,
						Icon:               &m.Icon,
						HideMenu:           &m.HideMenu,
						HideBreadcrumb:     &m.HideBreadcrumb,
						IgnoreKeepAlive:    &m.IgnoreKeepAlive,
						HideTab:            &m.HideTab,
						FrameSrc:           &m.FrameSrc,
						CarryParam:         &m.CarryParam,
						HideChildrenInMenu: &m.HideChildrenInMenu,
						Affix:              &m.Affix,
						DynamicLevel:       &m.DynamicLevel,
						RealPath:           &m.RealPath,
						Params:             &m.Params,
					},
				})
				existMap[m.ID] = struct{}{}
			}
		}
	}

	resp.Total = uint64(len(resp.Data))

	return resp, nil
}
