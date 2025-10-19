package oauthprovider

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/utils/jwt"
	"github.com/coder-lulu/newbee-common/utils/pointy"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type OauthCallbackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewOauthCallbackLogic(r *http.Request, svcCtx *svc.ServiceContext) *OauthCallbackLogic {
	return &OauthCallbackLogic{
		Logger: logx.WithContext(r.Context()),
		ctx:    r.Context(),
		svcCtx: svcCtx,
		r:      r,
	}
}

func (l *OauthCallbackLogic) OauthCallback() (resp *types.CallbackResp, err error) {
	result, err := l.svcCtx.CoreRpc.OauthCallback(l.ctx, &core.CallbackReq{
		State: l.r.FormValue("state"),
		Code:  l.r.FormValue("code"),
	})
	if err != nil {
		return nil, err
	}

	// Convert roleIds to string slice
	roleIdsStr := make([]string, len(result.RoleIds))
	for i, id := range result.RoleIds {
		roleIdsStr[i] = strconv.FormatUint(id, 10)
	}

	token, err := jwt.NewJwtToken(l.svcCtx.Config.Middleware.Auth.AccessSecret, time.Now().Unix(),
		l.svcCtx.Config.Middleware.Auth.AccessExpire,
		// 使用优化的短字段名
		jwt.WithOption(keys.JWTUserID, *result.Id),              // "uid"
		jwt.WithOption(keys.JWTTenantID, *result.TenantId),      // "tid"
		jwt.WithOption(keys.JWTUsername, *result.Username),      // "un"
		jwt.WithOption(keys.JWTDeptID, *result.DepartmentId),    // "did"
		jwt.WithOption(keys.JWTRoleCodes, strings.Join(result.RoleCodes, ",")), // "rc"
		// 用户信息
		jwt.WithOption(keys.JWTNickname, func() string {
			if result.Nickname != nil {
				return *result.Nickname
			}
			return ""
		}()),
		jwt.WithOption(keys.JWTAvatar, func() string {
			if result.Avatar != nil {
				return *result.Avatar
			}
			return ""
		}()),
		jwt.WithOption("tenantId", result.TenantId))

	// add token into database
	expiredAt := time.Now().Add(time.Second * time.Duration(l.svcCtx.Config.Middleware.Auth.AccessExpire)).UnixMilli()
	_, err = l.svcCtx.CoreRpc.CreateToken(l.ctx, &core.TokenInfo{
		Uuid:      result.Id,
		Token:     pointy.GetPointer(token),
		Source:    pointy.GetPointer(strings.Split(l.r.FormValue("state"), "-")[1]),
		Status:    pointy.GetPointer(uint32(1)),
		ExpiredAt: pointy.GetPointer(expiredAt),
		TenantId:  result.TenantId,
	})

	if err != nil {
		return nil, err
	}

	return &types.CallbackResp{
		UserId: *result.Id,
		Token:  token,
		Expire: uint64(expiredAt),
	}, nil
}
