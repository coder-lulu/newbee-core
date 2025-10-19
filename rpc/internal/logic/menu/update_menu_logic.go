package menu

import (
	"context"

	"github.com/coder-lulu/newbee-common/enum/common"
	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/menu"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuLogic {
	return &UpdateMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateMenuLogic) UpdateMenu(in *core.MenuInfo) (*core.BaseResp, error) {
	err := entx.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		// get parent level
		var menuLevel uint32
		if *in.ParentId != common.DefaultParentId {
			m, err := tx.Menu.Query().Where(menu.IDEQ(*in.ParentId)).First(l.ctx)
			if err != nil {
				return dberrorhandler.DefaultEntError(l.Logger, err, in)
			}

			menuLevel = m.MenuLevel + 1
		} else {
			menuLevel = 1
		}

		err := tx.Menu.UpdateOneID(*in.Id).
			SetNotNilMenuLevel(&menuLevel).
			SetNotNilMenuType(in.MenuType).
			SetNotNilParentID(in.ParentId).
			SetNotNilPath(in.Path).
			SetNotNilName(in.Name).
			SetNotNilRedirect(in.Redirect).
			SetNotNilComponent(in.Component).
			SetNotNilSort(in.Sort).
			SetNotNilDisabled(in.Disabled).
			SetNotNilServiceName(in.ServiceName).
			SetNotNilPermission(in.Permission).
			// meta
			SetNotNilTitle(in.Meta.Title).
			SetNotNilIcon(in.Meta.Icon).
			SetNotNilHideMenu(in.Meta.HideMenu).
			SetNotNilHideBreadcrumb(in.Meta.HideBreadcrumb).
			SetNotNilIgnoreKeepAlive(in.Meta.IgnoreKeepAlive).
			SetNotNilHideTab(in.Meta.HideTab).
			SetNotNilFrameSrc(in.Meta.FrameSrc).
			SetNotNilCarryParam(in.Meta.CarryParam).
			SetNotNilHideChildrenInMenu(in.Meta.HideChildrenInMenu).
			SetNotNilAffix(in.Meta.Affix).
			SetNotNilDynamicLevel(in.Meta.DynamicLevel).
			SetNotNilRealPath(in.Meta.RealPath).
			SetNotNilParams(in.Meta.Params).
			Exec(l.ctx)
		if err != nil {
			return dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		if *in.Disabled {
			menu, err := tx.Menu.Query().Where(menu.ID(*in.Id)).WithChildren().First(l.ctx)
			if err != nil {
				return dberrorhandler.DefaultEntError(l.Logger, err, in)

			}
			for _, v := range menu.Edges.Children {
				err = tx.Menu.UpdateOneID(v.ID).SetNotNilDisabled(in.Disabled).Exec(l.ctx)
				if err != nil {
					return dberrorhandler.DefaultEntError(l.Logger, err, in)
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
