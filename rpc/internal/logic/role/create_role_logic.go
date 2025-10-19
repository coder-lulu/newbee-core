package role

import (
	"context"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"

	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/i18n"
)

type CreateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateRoleLogic {
	return &CreateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateRoleLogic) CreateRole(in *core.RoleInfo) (*core.BaseIDResp, error) {
	err := entx.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		result, err := tx.Role.Create().
			SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
			SetNotNilName(in.Name).
			SetNotNilCode(in.Code).
			SetNotNilDefaultRouter(in.DefaultRouter).
			SetNotNilRemark(in.Remark).
			SetNotNilSort(in.Sort).
			// 🔥 Phase 3: data_scope field removed - now managed via sys_casbin_rules
			SetNotNilCustomDeptIds(in.CustomDeptIds).
			Save(l.ctx)
		if err != nil {
			return err
		}

		// 菜单授权。原authority接口不再调用
		err = tx.Role.UpdateOneID(result.ID).ClearMenus().Exec(l.ctx)
		if err != nil {
			return err
		}

		err = tx.Role.UpdateOneID(result.ID).AddMenuIDs(in.MenuIds...).Exec(l.ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	return &core.BaseIDResp{Id: 0, Msg: i18n.CreateSuccess}, nil
}
