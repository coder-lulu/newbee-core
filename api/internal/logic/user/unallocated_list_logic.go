package user

import (
	"context"
	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnallocatedListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnallocatedListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnallocatedListLogic {
	return &UnallocatedListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *UnallocatedListLogic) UnallocatedList(req *types.RoleUnallocatedUserListReq) (resp *types.UserListResp, err error) {
	data, err := l.svcCtx.CoreRpc.UnallocatedList(l.ctx, &core.RoleUnallocatedListReq{
		Page:     req.Page,
		PageSize: req.PageSize,
		RoleId:   req.RoleId,
		UserName: req.UserName,
		Mobile:   req.Mobile,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.UserListResp{}
	for _, v := range data.Data {
		resp.Data.Data = append(resp.Data.Data, types.UserInfo{
			BaseUUIDInfo: types.BaseUUIDInfo{
				Id: v.Id,
			},
			Username:    v.Username,
			Nickname:    v.Nickname,
			Mobile:      v.Mobile,
			Email:       v.Email,
			Avatar:      v.Avatar,
			Status:      v.Status,
			Description: v.Description,
		})
	}
	resp.Data.Total = data.Total
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, i18n.Success)
	return resp, nil
}
