package user

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/utils/encrypt"
	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent/user"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/coder-lulu/newbee-common/v2/i18n"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *core.UserInfo) (*core.BaseUUIDResp, error) {
	if in.Mobile != nil {
		checkMobile, err := l.svcCtx.DB.User.Query().Where(user.MobileEQ(*in.Mobile)).Exist(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if checkMobile {
			return nil, errorx.NewInvalidArgumentError("login.mobileExist")
		}
	}

	if in.Email != nil {
		checkEmail, err := l.svcCtx.DB.User.Query().Where(user.EmailEQ(*in.Email)).Exist(l.ctx)
		if err != nil {
			return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
		}

		if checkEmail {
			return nil, errorx.NewInvalidArgumentError("login.signupUserExist")
		}
	}

	result, err := l.svcCtx.DB.User.Create().
		SetNotNilUsername(in.Username).
		SetNotNilPassword(pointy.GetPointer(encrypt.BcryptEncrypt(*in.Password))).
		SetNotNilNickname(in.Nickname).
		SetNotNilEmail(in.Email).
		SetNotNilMobile(in.Mobile).
		SetNotNilAvatar(in.Avatar).
		AddRoleIDs(in.RoleIds...).
		SetNotNilHomePath(in.HomePath).
		SetNotNilDescription(in.Description).
		SetNotNilDepartmentID(in.DepartmentId).
		AddPositionIDs(in.PositionIds...).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseUUIDResp{Id: result.ID.String(), Msg: i18n.CreateSuccess}, nil
}
