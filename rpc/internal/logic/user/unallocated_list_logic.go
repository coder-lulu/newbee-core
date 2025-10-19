package user

import (
	"context"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"
	"github.com/coder-lulu/newbee-core/rpc/ent/role"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnallocatedListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnallocatedListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnallocatedListLogic {
	return &UnallocatedListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnallocatedListLogic) UnallocatedList(in *core.RoleUnallocatedListReq) (*core.UserListResp, error) {
	var predicates []predicate.User

	if in.Mobile != nil {
		predicates = append(predicates, user.MobileEQ(*in.Mobile))
	}

	if in.UserName != nil {
		predicates = append(predicates, user.UsernameContains(*in.UserName))
	}
	// 排除已经授权的用户
	allocatedUserIds, err := l.svcCtx.DB.User.Query().Where(user.HasRolesWith(role.ID(in.RoleId))).IDs(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}
	predicates = append(predicates, user.IDNotIn(allocatedUserIds...))
	users, err := l.svcCtx.DB.User.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)
	fmt.Println(users)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.UserListResp{}
	resp.Total = users.PageDetails.Total

	for _, v := range users.List {
		resp.Data = append(resp.Data, &core.UserInfo{
			Id:          pointy.GetPointer(v.ID.String()),
			Avatar:      &v.Avatar,
			Mobile:      &v.Mobile,
			Email:       &v.Email,
			Status:      pointy.GetPointer(uint32(v.Status)),
			Username:    &v.Username,
			Nickname:    &v.Nickname,
			Description: &v.Description,
		})
	}

	return resp, nil
}
