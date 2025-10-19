package user

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/coder-lulu/newbee-common/v2/orm/ent/entctx/userctx"
	"github.com/coder-lulu/newbee-common/v2/utils/jwt"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *RefreshTokenLogic) RefreshToken() (resp *types.RefreshTokenResp, err error) {
	userId, err := userctx.GetUserIDFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	userData, err := l.svcCtx.CoreRpc.GetUserById(l.ctx, &core.UUIDReq{
		Id: userId,
	})
	if err != nil {
		// Check if it's an authentication error (deleted user)
		if strings.Contains(err.Error(), "Token is invalid") {
			return nil, errorx.NewCodeError(http.StatusUnauthorized, "Token is invalid")
		}
		return nil, err
	}

	if userData.Status != nil && *userData.Status != uint32(common.StatusNormal) {
		return nil, errorx.NewApiUnauthorizedError(i18n.Failed)
	}

	// Convert roleIds to string slice
	roleIdsStr := make([]string, len(userData.RoleIds))
	for i, id := range userData.RoleIds {
		roleIdsStr[i] = strconv.FormatUint(id, 10)
	}

	token, err := jwt.NewJwtToken(l.svcCtx.Config.Middleware.Auth.AccessSecret, time.Now().Unix(),
		int64(l.svcCtx.Config.ProjectConf.RefreshTokenPeriod)*60*60,
		// 使用优化的短字段名
		jwt.WithOption(keys.JWTUserID, userId),                      // "uid"
		jwt.WithOption(keys.JWTTenantID, *userData.TenantId),         // "tid"
		jwt.WithOption(keys.JWTUsername, *userData.Username),         // "un"
		jwt.WithOption(keys.JWTDeptID, *userData.DepartmentId),       // "did"
		jwt.WithOption(keys.JWTRoleCodes, strings.Join(userData.RoleCodes, ",")), // "rc"
		// 用户信息
		jwt.WithOption(keys.JWTNickname, func() string {
			if userData.Nickname != nil {
				return *userData.Nickname
			}
			return ""
		}()),
		jwt.WithOption(keys.JWTAvatar, func() string {
			if userData.Avatar != nil {
				return *userData.Avatar
			}
			return ""
		}()))
	if err != nil {
		return nil, err
	}

	// add token into database
	expiredAt := time.Now().Add(time.Hour * time.Duration(l.svcCtx.Config.ProjectConf.RefreshTokenPeriod)).UnixMilli()
	_, err = l.svcCtx.CoreRpc.CreateToken(l.ctx, &core.TokenInfo{
		Uuid:      &userId,
		Token:     pointy.GetPointer(token),
		Source:    pointy.GetPointer("core_user_refresh_token"),
		Status:    pointy.GetPointer(uint32(common.StatusNormal)),
		Username:  userData.Username,
		ExpiredAt: pointy.GetPointer(expiredAt),
	})

	return &types.RefreshTokenResp{
		BaseDataInfo: types.BaseDataInfo{Msg: i18n.Success},
		Data:         types.RefreshTokenInfo{Token: token, ExpiredAt: expiredAt},
	}, nil
}
