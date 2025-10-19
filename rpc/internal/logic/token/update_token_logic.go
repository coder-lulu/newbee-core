package token

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-common/v2/config"

	"github.com/coder-lulu/newbee-common/v2/enum/common"
	"github.com/coder-lulu/newbee-common/v2/i18n"
	"github.com/coder-lulu/newbee-common/v2/msg/logmsg"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"
	"github.com/coder-lulu/newbee-common/v2/utils/uuidx"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTokenLogic {
	return &UpdateTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTokenLogic) UpdateToken(in *core.TokenInfo) (*core.BaseResp, error) {
	token, err := l.svcCtx.DB.Token.UpdateOneID(uuidx.ParseUUIDString(*in.Id)).
		SetNotNilStatus(pointy.GetStatusPointer(in.Status)).
		SetNotNilSource(in.Source).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if uint8(*in.Status) == common.StatusBanned {
		expiredTime := time.Until(token.ExpiredAt)
		if expiredTime > 0 {
			err = l.svcCtx.Redis.Set(l.ctx, config.RedisTokenPrefix+token.Token, "1", expiredTime).Err()
			if err != nil {
				logx.Errorw(logmsg.RedisError, logx.Field("detail", err.Error()))
				return nil, errorx.NewInternalError(i18n.RedisError)
			}
		}
	} else if uint8(*in.Status) == common.StatusNormal {
		err := l.svcCtx.Redis.Del(l.ctx, config.RedisTokenPrefix+token.Token).Err()
		if err != nil {
			logx.Errorw(logmsg.RedisError, logx.Field("detail", err.Error()))
			return nil, errorx.NewInternalError(i18n.RedisError)
		}
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
