package user

import (
	"context"
	"net/http"
	"strings"

	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/middleware/keys"
	"github.com/zeromicro/go-zero/core/errorx"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserPermCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserPermCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserPermCodeLogic {
	return &GetUserPermCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserPermCodeLogic) GetUserPermCode() (resp *types.PermCodeResp, err error) {
	// 从context获取角色编码
	roleCodes := ""
	if val := l.ctx.Value(keys.RoleCodesKey); val != nil {
		if codes, ok := val.(string); ok {
			roleCodes = codes
		}
	}
	if roleCodes == "" {
		return nil, &errorx.ApiError{
			Code: http.StatusUnauthorized,
			Msg:  "login.requireLogin",
		}
	}

	return &types.PermCodeResp{
		BaseDataInfo: types.BaseDataInfo{Msg: l.svcCtx.Trans.Trans(l.ctx, i18n.Success)},
		Data:         strings.Split(roleCodes, ","),
	}, nil
}
