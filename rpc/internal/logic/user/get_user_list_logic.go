package user

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/datapermctx"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent/department"
	"github.com/coder-lulu/newbee-core/rpc/ent/position"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserListLogic) GetUserList(in *core.UserListReq) (*core.UserListResp, error) {
	var predicates []predicate.User

	if in.Mobile != nil {
		predicates = append(predicates, user.MobileEQ(*in.Mobile))
	}

	if in.Username != nil {
		predicates = append(predicates, user.UsernameContains(*in.Username))
	}

	if in.Email != nil {
		predicates = append(predicates, user.EmailEQ(*in.Email))
	}

	if in.Nickname != nil {
		predicates = append(predicates, user.NicknameContains(*in.Nickname))
	}

	if in.RoleIds != nil {
		predicates = append(predicates, user.HasRolesWith(role.IDIn(in.RoleIds...)))
	}

	if in.DepartmentId != nil {
		var lists []uint64
		queue := []uint64{*in.DepartmentId} // 使用队列代替递归

		for len(queue) > 0 {
			currentDeptId := queue[0]
			queue = queue[1:]

			// 添加当前部门 ID
			lists = append(lists, currentDeptId)

			// 查询当前部门的子部门
			result, err := l.svcCtx.DB.Department.Query().
				Where(department.ParentID(currentDeptId)).
				All(l.ctx)
			if err != nil {
				return nil, err
			}

			// 将子部门 ID 加入队列
			for _, v := range result {
				queue = append(queue, v.ID)
			}
		}

		predicates = append(predicates, user.DepartmentIDIn(lists...))
	}

	if in.PositionIds != nil {
		predicates = append(predicates, user.HasPositionsWith(position.IDIn(in.PositionIds...)))
	}

	if in.Description != nil {
		predicates = append(predicates, user.DescriptionContains(*in.Description))
	}

	// filter only user own
	l.ctx = datapermctx.WithFilterFieldContext(l.ctx, user.FieldID)

	users, err := l.svcCtx.DB.User.Query().Where(predicates...).WithRoles().WithPositions().Page(l.ctx, in.Page, in.PageSize)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.UserListResp{}
	resp.Total = users.PageDetails.Total

	for _, v := range users.List {
		dept, _ := l.svcCtx.DB.Department.Query().Where(department.ID(v.DepartmentID)).First(l.ctx)
		deptName := "无部门"
		if dept != nil {
			deptName = dept.Name
		}
		resp.Data = append(resp.Data, &core.UserInfo{
			Id:             pointy.GetPointer(v.ID.String()),
			Avatar:         &v.Avatar,
			RoleIds:        GetRoleIds(v.Edges.Roles),
			RoleCodes:      GetRoleCodes(v.Edges.Roles),
			Mobile:         &v.Mobile,
			Email:          &v.Email,
			Status:         pointy.GetPointer(uint32(v.Status)),
			Username:       &v.Username,
			Nickname:       &v.Nickname,
			HomePath:       &v.HomePath,
			Description:    &v.Description,
			DepartmentId:   &v.DepartmentID,
			DepartmentName: &deptName,
			PositionIds:    GetPositionIds(v.Edges.Positions),
			CreatedAt:      pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:      pointy.GetPointer(v.UpdatedAt.UnixMilli()),
		})
	}

	return resp, nil
}
