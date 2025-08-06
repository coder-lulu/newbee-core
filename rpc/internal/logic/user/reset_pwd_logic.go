package user

import (
	"context"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/utils/encrypt"
	"github.com/coder-lulu/newbee-common/utils/uuidx"
	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPwdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPwdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPwdLogic {
	return &ResetPwdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetPwdLogic) ResetPwd(in *core.ResetPwdReq) (*core.BaseResp, error) {
	if in.OpId != nil {
		user, err := l.svcCtx.DB.User.Query().Where(user.ID(uuidx.ParseUUIDString(*in.OpId))).WithRoles().First(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}
		if !IsSuperAdmin(user.Edges.Roles) {
			return &core.BaseResp{Msg: "您无此权限"}, nil

		}
	} else {
		return &core.BaseResp{Msg: "您无此权限"}, nil
	}
	err := l.svcCtx.DB.User.UpdateOneID(uuidx.ParseUUIDString(in.UserId)).
		SetPassword(encrypt.BcryptEncrypt(in.Password)).Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
