package department

import (
	"context"
	"fmt"
	"github.com/coder-lulu/newbee-core/rpc/ent/position"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/datapermctx"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entenum"

	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/redisfunc"

	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/ent/department"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/v2/config"
	"github.com/coder-lulu/newbee-common/v2/i18n"
)

type DeleteDepartmentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteDepartmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDepartmentLogic {
	return &DeleteDepartmentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func contains(slice []uint64, target uint64) bool {
	for _, value := range slice {
		if value == target {
			return true
		}
	}
	return false
}

func (l *DeleteDepartmentLogic) DeleteDepartment(in *core.IDsReq) (*core.BaseResp, error) {
	if contains(in.Ids, 1) {
		logx.Errorw("delete department failed, the default department can not be deleted",
			logx.Field("departmentId", in.Ids))
		return nil, errorx.NewInvalidArgumentError("默认部门无法被删除")

	}
	exist, err := l.svcCtx.DB.Department.Query().Where(department.ParentIDIn(in.Ids...)).Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if exist {
		logx.Errorw("delete department failed, please check its children had been deleted",
			logx.Field("departmentId", in.Ids))
		return nil, errorx.NewInvalidArgumentError("department.deleteDepartmentChildrenFirst")
	}

	checkPosts, err := l.svcCtx.DB.Position.Query().Where(position.DeptIDIn(in.Ids...)).Exist(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if checkPosts {
		logx.Errorw("delete department failed, there are posts belongs to the department", logx.Field("departmentId", in.Ids))
		return nil, errorx.NewInvalidArgumentError("删除失败, 部门下存在所属岗位关联")
	}

	checkUser, err := l.svcCtx.DB.User.Query().Where(user.DepartmentIDIn(in.Ids...)).Exist(datapermctx.WithScopeContext(l.ctx, entenum.DataPermAllStr))
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	if checkUser {
		logx.Errorw("delete department failed, there are users belongs to the department", logx.Field("departmentId", in.Ids))
		return nil, errorx.NewInvalidArgumentError("department.deleteDepartmentUserFirst")
	}

	_, err = l.svcCtx.DB.Department.Delete().Where(department.IDIn(in.Ids...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	err = redisfunc.RemoveAllKeyByPrefix(l.ctx, fmt.Sprintf("%sDEPT", config.RedisDataPermissionPrefix), l.svcCtx.Redis)
	if err != nil {
		return nil, err
	}

	return &core.BaseResp{Msg: i18n.DeleteSuccess}, nil
}
