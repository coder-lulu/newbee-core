package publicuser

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/config"
	"github.com/coder-lulu/newbee-common/enum/common"
	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/coder-lulu/newbee-common/orm/ent/entctx/datapermctx"
	"github.com/coder-lulu/newbee-common/orm/ent/entenum"
	"github.com/coder-lulu/newbee-common/utils/jwt"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginByEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginByEmailLogic {
	return &LoginByEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx}
}

func (l *LoginByEmailLogic) LoginByEmail(req *types.LoginByEmailReq) (resp *types.LoginResp, err error) {
	if l.svcCtx.Config.ProjectConf.LoginVerify != "email" && l.svcCtx.Config.ProjectConf.LoginVerify != "sms_or_email" &&
		l.svcCtx.Config.ProjectConf.LoginVerify != "all" {
		return nil, errorx.NewCodeAbortedError("login.loginTypeForbidden")
	}

	captchaData, err := l.svcCtx.Redis.Get(l.ctx, config.RedisCaptchaPrefix+req.Email).Result()
	if err != nil {
		logx.Errorw("failed to get captcha data in redis for email validation", logx.Field("detail", err),
			logx.Field("data", req))
		return nil, errorx.NewCodeInvalidArgumentError(i18n.Failed)
	}

	if captchaData == req.Captcha {
		l.ctx = datapermctx.WithScopeContext(l.ctx, entenum.DataPermAllStr)

		userData, err := l.svcCtx.CoreRpc.GetUserList(l.ctx, &core.UserListReq{
			Page:     1,
			PageSize: 1,
			Email:    &req.Email,
		})
		if err != nil {
			return nil, err
		}

		if userData.Total == 0 {
			return nil, errorx.NewCodeInvalidArgumentError("login.userNotExist")
		}

		if *userData.Data[0].Status != uint32(common.StatusNormal) {
			return nil, errorx.NewCodeInvalidArgumentError("login.userBanned")
		}

		// Convert roleIds to string slice
		roleIdsStr := make([]string, len(userData.Data[0].RoleIds))
		for i, id := range userData.Data[0].RoleIds {
			roleIdsStr[i] = strconv.FormatUint(id, 10)
		}

		token, err := jwt.NewJwtToken(l.svcCtx.Config.Middleware.Auth.AccessSecret, time.Now().Unix(),
			l.svcCtx.Config.Middleware.Auth.AccessExpire,
			// 使用优化的短字段名
			jwt.WithOption(keys.JWTUserID, *userData.Data[0].Id),           // "uid"
			jwt.WithOption(keys.JWTTenantID, *userData.Data[0].TenantId),   // "tid"
			jwt.WithOption(keys.JWTUsername, *userData.Data[0].Username),   // "un"
			jwt.WithOption(keys.JWTDeptID, *userData.Data[0].DepartmentId), // "did"
			jwt.WithOption(keys.JWTRoleCodes, strings.Join(userData.Data[0].RoleCodes, ",")), // "rc"
			// 用户信息
			jwt.WithOption(keys.JWTNickname, func() string {
				if userData.Data[0].Nickname != nil {
					return *userData.Data[0].Nickname
				}
				return ""
			}()),
			jwt.WithOption(keys.JWTAvatar, func() string {
				if userData.Data[0].Avatar != nil {
					return *userData.Data[0].Avatar
				}
				return ""
			}()))
		if err != nil {
			return nil, err
		}

		// add token into database
		expiredAt := time.Now().Add(time.Second * time.Duration(l.svcCtx.Config.Middleware.Auth.AccessExpire)).UnixMilli()
		_, err = l.svcCtx.CoreRpc.CreateToken(l.ctx, &core.TokenInfo{
			Uuid:      userData.Data[0].Id,
			Token:     pointy.GetPointer(token),
			Source:    pointy.GetPointer("core_user"),
			Status:    pointy.GetPointer(uint32(common.StatusNormal)),
			Username:  userData.Data[0].Username,
			ExpiredAt: pointy.GetPointer(expiredAt),
			TenantId:  userData.Data[0].TenantId,
		})

		if err != nil {
			return nil, err
		}

		err = l.svcCtx.Redis.Del(l.ctx, config.RedisCaptchaPrefix+req.Email).Err()
		if err != nil {
			logx.Errorw("failed to delete captcha in redis", logx.Field("detail", err))
		}

		resp = &types.LoginResp{
			BaseDataInfo: types.BaseDataInfo{Msg: l.svcCtx.Trans.Trans(l.ctx, "login.loginSuccessTitle")},
			Data: types.LoginInfo{
				UserId: *userData.Data[0].Id,
				Token:  token,
				Expire: uint64(expiredAt),
			},
		}
		return resp, nil
	} else {
		return nil, errorx.NewCodeInvalidArgumentError("login.wrongCaptcha")
	}
}
