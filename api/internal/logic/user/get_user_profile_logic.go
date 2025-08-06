package user

import (
	"context"
	"github.com/coder-lulu/newbee-common/i18n"
	"strings"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileLogic {
	return &GetUserProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserProfileLogic) GetUserProfile() (resp *types.ProfileResp, err error) {
	data, err := l.svcCtx.CoreRpc.GetUserById(l.ctx, &core.UUIDReq{Id: l.ctx.Value("userId").(string)})
	if err != nil {
		return nil, err
	}

	user := types.UserInfo{
		BaseUUIDInfo: types.BaseUUIDInfo{
			Id:        data.Id,
			CreatedAt: data.CreatedAt,
		},
		Status:         data.Status,
		Username:       data.Username,
		Nickname:       data.Nickname,
		Description:    data.Description,
		HomePath:       data.HomePath,
		Mobile:         data.Mobile,
		Email:          data.Email,
		Avatar:         data.Avatar,
		DepartmentName: data.DepartmentName,
	}
	roleName := strings.Join(data.RoleNames, ",")
	postIds := data.GetPositionIds()
	var posts []string
	for _, pid := range postIds {
		p, err := l.svcCtx.CoreRpc.GetPositionById(l.ctx, &core.IDReq{
			Id: pid,
		})
		if err == nil && p != nil && p.Name != nil {
			posts = append(posts, *p.Name)
		}
	}
	postName := strings.Join(posts, ",")
	return &types.ProfileResp{
		BaseDataInfo: types.BaseDataInfo{Msg: l.svcCtx.Trans.Trans(l.ctx, i18n.Success)},
		Data: types.ProfileUserInfo{
			User:      user,
			RoleGroup: &roleName,
			PostGroup: &postName,
		},
	}, nil
}
