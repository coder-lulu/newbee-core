package token

import (
	"context"
	"time"

	"github.com/coder-lulu/newbee-common/config"
	"github.com/coder-lulu/newbee-common/msg/logmsg"
	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/i18n"

	"github.com/coder-lulu/newbee-core/rpc/ent/token"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
)

type DeleteTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTokenLogic {
	return &DeleteTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteTokenLogic) DeleteToken(in *core.UUIDsReq) (*core.BaseResp, error) {
	_, err := l.svcCtx.DB.Token.Delete().Where(token.IDIn(uuidx.ParseUUIDSlice(in.Ids)...)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 删除token后将token加入黑名单
	for _, v := range in.Ids {
		tokenData, err := l.svcCtx.DB.Token.Query().Where(token.ID(uuidx.ParseUUIDString(v))).First(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		err = l.svcCtx.Redis.Set(l.ctx, config.RedisTokenPrefix+v, "1", time.Until(tokenData.ExpiredAt)).Err()
		if err != nil {
			logx.Errorw(logmsg.RedisError, logx.Field("detail", err.Error()))
			return nil, errorx.NewInternalError(i18n.RedisError)
		}
	}

	return &core.BaseResp{Msg: i18n.DeleteSuccess}, nil
}
