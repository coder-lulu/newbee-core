package publicuser

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/coder-lulu/newbee-common/v2/config"
	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/middleware/keys"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/coder-lulu/newbee-common/v2/utils/encrypt"
	"github.com/coder-lulu/newbee-common/v2/utils/jwt"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	if l.svcCtx.Config.ProjectConf.LoginVerify != "captcha" && l.svcCtx.Config.ProjectConf.LoginVerify != "all" {
		return nil, errorx.NewCodeAbortedError("login.loginTypeForbidden")
	}

	if ok := l.svcCtx.Captcha.Verify(config.RedisCaptchaPrefix+req.CaptchaId, req.Captcha, true); ok {

		// check if input wrong password too many times, forbidden for 5 minutes
		checkWrongTimes := 0
		if checkWrongTimesStr, err := l.svcCtx.Redis.Get(l.ctx, "USER:WRONG_PASSWORD:"+req.Username).Result(); err != nil {
			if !errors.Is(err, redis.Nil) {
				return nil, err
			}
		} else {
			checkWrongTimes, err = strconv.Atoi(checkWrongTimesStr)
			if err != nil {
				return nil, err
			}

			if checkWrongTimes > 5 {
				return nil, errorx.NewCodeAbortedError("login.wrongPasswordOverTimes")
			}
		}

		tenantIdUint, err := strconv.ParseUint(req.TenantId, 10, 64)
		if err != nil {
			return nil, errorx.NewCodeInvalidArgumentError("login.invalidTenant")
		}
		tenantInfo, err := l.svcCtx.CoreRpc.GetTenantById(l.ctx, &core.IDReq{Id: tenantIdUint})
		logx.Info("租户信息：%v", tenantInfo)
		if err != nil {
			if e, ok := status.FromError(err); ok {
				if e.Message() == i18n.TargetNotFound {
					return nil, errorx.NewCodeInvalidArgumentError("login.invalidTenant")
				}
			}

			return nil, err
		}

		if tenantInfo.Status != nil && *tenantInfo.Status != uint32(common.StatusNormal) {
			return nil, errorx.NewCodeInvalidArgumentError("login.tenantDisabled")
		}

		tenantID := tenantInfo.GetId()
		if tenantID == 0 {
			return nil, errorx.NewCodeInvalidArgumentError("login.invalidTenant")
		}
		tenantIDStr := strconv.FormatUint(tenantID, 10)
		tenantCtx := l.svcCtx.ContextManager.SetTenantID(l.ctx, tenantIDStr)
		tenantCtx = metadata.AppendToOutgoingContext(tenantCtx, keys.TenantIDKey.String(), tenantIDStr)
		l.ctx = tenantCtx

		user, err := l.svcCtx.CoreRpc.GetUserByUsername(l.ctx,
			&core.UsernameReq{
				Username: req.Username,
			})
		if err != nil {
			if e, ok := status.FromError(err); ok {
				if e.Message() == i18n.TargetNotFound {
					return nil, errorx.NewCodeInvalidArgumentError("login.wrongUsernameOrPassword")
				}
			}

			return nil, err
		}

		if user.Status != nil && *user.Status != uint32(common.StatusNormal) {
			return nil, errorx.NewCodeInvalidArgumentError("login.userBanned")
		}

		if !encrypt.BcryptCheck(req.Password, *user.Password) {
			if err = l.svcCtx.Redis.Set(l.ctx, "USER:WRONG_PASSWORD:"+req.Username, checkWrongTimes+1, 5*time.Minute).Err(); err != nil {
				return nil, err
			}

			return nil, errorx.NewCodeInvalidArgumentError("login.wrongUsernameOrPassword")
		}

		// Convert roleIds to string slice
		roleIdsStr := make([]string, len(user.RoleIds))
		for i, id := range user.RoleIds {
			roleIdsStr[i] = strconv.FormatUint(id, 10)
		}

		token, err := jwt.NewJwtToken(l.svcCtx.Config.Middleware.Auth.AccessSecret, time.Now().Unix(),
			l.svcCtx.Config.Middleware.Auth.AccessExpire,
			// 使用优化的短字段名，减少token长度
			jwt.WithOption(keys.JWTUserID, *user.Id),                             // "uid"
			jwt.WithOption(keys.JWTTenantID, *user.TenantId),                     // "tid"
			jwt.WithOption(keys.JWTUsername, *user.Username),                     // "un"
			jwt.WithOption(keys.JWTDeptID, *user.DepartmentId),                   // "did"
			jwt.WithOption(keys.JWTRoleCodes, strings.Join(user.RoleCodes, ",")), // "rc"
			// 用户信息 - 可选字段
			jwt.WithOption(keys.JWTNickname, func() string { // "nn"
				if user.Nickname != nil {
					return *user.Nickname
				}
				return ""
			}()),
			jwt.WithOption(keys.JWTAvatar, func() string { // "av"
				if user.Avatar != nil {
					return *user.Avatar
				}
				return ""
			}()))
		if err != nil {
			return nil, err
		}

		// add token into database
		expiredAt := time.Now().Add(time.Second * time.Duration(l.svcCtx.Config.Middleware.Auth.AccessExpire)).UnixMilli()
		_, err = l.svcCtx.CoreRpc.CreateToken(l.ctx, &core.TokenInfo{
			Uuid:      user.Id,
			Token:     pointy.GetPointer(token),
			Source:    pointy.GetPointer("core_user"),
			Status:    pointy.GetPointer(uint32(common.StatusNormal)),
			Username:  user.Username,
			ExpiredAt: pointy.GetPointer(expiredAt),
			TenantId:  user.TenantId,
		})

		if err != nil {
			return nil, err
		}

		err = l.svcCtx.Redis.Del(l.ctx, config.RedisCaptchaPrefix+req.CaptchaId).Err()
		if err != nil {
			logx.Errorw("failed to delete captcha in redis", logx.Field("detail", err))
		}

		resp = &types.LoginResp{
			BaseDataInfo: types.BaseDataInfo{Msg: l.svcCtx.Trans.Trans(l.ctx, "login.loginSuccessTitle")},
			Data: types.LoginInfo{
				UserId: *user.Id,
				Token:  token,
				Expire: uint64(expiredAt),
			},
		}
		return resp, nil
	} else {
		return nil, errorx.NewCodeInvalidArgumentError("login.wrongCaptcha")
	}
}
