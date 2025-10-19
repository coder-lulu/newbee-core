package user

import (
	"context"
	"net/http"
	"strings"

	"github.com/coder-lulu/newbee-common/utils/encrypt"
	"github.com/coder-lulu/newbee-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type ChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangePasswordLogic {
	return &ChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangePasswordLogic) ChangePassword(req *types.ChangePasswordReq) (resp *types.BaseMsgResp, err error) {
	userData, err := l.svcCtx.CoreRpc.GetUserById(l.ctx, &core.UUIDReq{Id: l.svcCtx.ContextManager.GetUserID(l.ctx)})
	if err != nil {
		// Check if it's an authentication error (deleted user)
		if strings.Contains(err.Error(), "Token is invalid") {
			return nil, errorx.NewCodeError(http.StatusUnauthorized, "Token is invalid")
		}
		return nil, err
	}

	if encrypt.BcryptCheck(req.OldPassword, *userData.Password) {
		result, err := l.svcCtx.CoreRpc.UpdateUser(l.ctx, &core.UserInfo{
			Id:       pointy.GetPointer(l.svcCtx.ContextManager.GetUserID(l.ctx)),
			Password: pointy.GetPointer(req.NewPassword),
		})
		if err != nil {
			return nil, err
		}

		return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, result.Msg)}, nil
	}

	return nil, errorx.NewCodeInvalidArgumentError("login.wrongPassword")
}
