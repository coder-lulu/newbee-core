package role

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/config"

	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"

	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/entx"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/v2/i18n"
)

type UpdateRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRoleLogic {
	return &UpdateRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateRoleLogic) UpdateRole(in *core.RoleInfo) (*core.BaseResp, error) {
	err := entx.WithTx(l.ctx, l.svcCtx.DB, func(tx *ent.Tx) error {
		origin, err := tx.Role.Get(l.ctx, *in.Id)
		if err != nil {
			return err
		}

		err = tx.Role.UpdateOneID(*in.Id).
			SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
			SetNotNilName(in.Name).
			SetNotNilDefaultRouter(in.DefaultRouter).
			SetNotNilRemark(in.Remark).
			SetNotNilSort(in.Sort).
			// 🔥 Phase 3: data_scope field removed - now managed via sys_casbin_rules
			SetNotNilCustomDeptIds(in.CustomDeptIds).
			Exec(l.ctx)

		if err != nil {
			return err
		}

		if in.Code != nil && origin.Code != *in.Code {
			_, err = tx.QueryContext(l.ctx, fmt.Sprintf("update casbin_rules set v0='%s' WHERE v0='%s'", *in.Code, origin.Code))
			if err != nil {
				return err
			}
		}
		// 菜单授权。原authority接口不再调用
		err = tx.Role.UpdateOneID(*in.Id).ClearMenus().Exec(l.ctx)
		if err != nil {
			return err
		}
		err = tx.Role.UpdateOneID(*in.Id).AddMenuIDs(in.MenuIds...).Exec(l.ctx)
		if err != nil {
			return err
		}

		err = redisfunc.RemoveAllKeyByPrefix(l.ctx, fmt.Sprintf("%sROLE", config.RedisDataPermissionPrefix), l.svcCtx.Redis)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
