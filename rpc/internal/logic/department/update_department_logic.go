package department

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/config"

	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"

	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dbfunc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/i18n"
)

type UpdateDepartmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDepartmentLogic {
	return &UpdateDepartmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDepartmentLogic) UpdateDepartment(in *core.DepartmentInfo) (*core.BaseResp, error) {

	ancestors, err := dbfunc.GetDepartmentAncestors(in.ParentId, l.svcCtx.DB, l.Logger, l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = l.svcCtx.DB.Department.UpdateOneID(*in.Id).
		SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
		SetNotNilSort(in.Sort).
		SetNotNilName(in.Name).
		SetNotNilAncestors(ancestors).
		SetNotNilLeader(in.Leader).
		SetNotNilPhone(in.Phone).
		SetNotNilEmail(in.Email).
		SetNotNilRemark(in.Remark).
		SetNotNilParentID(in.ParentId).
		Exec(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = redisfunc.RemoveAllKeyByPrefix(l.ctx, fmt.Sprintf("%sDEPT", config.RedisDataPermissionPrefix), l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
