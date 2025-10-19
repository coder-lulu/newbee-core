package user

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo() (resp *types.UserBaseIDInfoResp, err error) {
	// get user id from context
	userId := l.svcCtx.ContextManager.GetUserID(l.ctx)
	l.Infof("API GetUserInfo start - userId: %s", userId)
	rpcStart := time.Now()

	user, err := l.svcCtx.CoreRpc.GetUserById(l.ctx,
		&core.UUIDReq{Id: userId})
	rpcDuration := time.Since(rpcStart)
	l.Infof("API GetUserInfo RPC call completed - duration: %v", rpcDuration)

	if err != nil {
		l.Errorf("API GetUserInfo RPC call error - duration: %v, error: %v", rpcDuration, err)
		// Check if it's an authentication error (deleted user)
		if strings.Contains(err.Error(), "Token is invalid") {
			return nil, errorx.NewCodeError(http.StatusUnauthorized, "Token is invalid")
		}
		return nil, err
	}

	return &types.UserBaseIDInfoResp{
		BaseDataInfo: types.BaseDataInfo{Msg: l.svcCtx.Trans.Trans(l.ctx, i18n.Success)},
		Data: types.UserBaseIDInfo{
			UUID:           user.Id,
			Username:       user.Username,
			Nickname:       user.Nickname,
			Avatar:         user.Avatar,
			HomePath:       user.HomePath,
			Description:    user.Description,
			DepartmentName: l.svcCtx.Trans.Trans(l.ctx, *user.DepartmentName),
			RoleNames:      TransRoleName(l.svcCtx, l.ctx, user.RoleNames),
			RoleCodes:      user.RoleCodes,
		},
	}, nil
}

// TransRoleName returns the i18n translation of role name slice.
func TransRoleName(svc *svc.ServiceContext, ctx context.Context, data []string) []string {
	var result []string
	for _, v := range data {
		result = append(result, svc.Trans.Trans(ctx, v))
	}
	return result
}
