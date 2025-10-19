package oauthaccount

import (
	"context"

	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserOauthAccountsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserOauthAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOauthAccountsLogic {
	return &GetUserOauthAccountsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserOauthAccountsLogic) GetUserOauthAccounts(req *types.GetUserOauthAccountsReq) (resp *types.GetUserOauthAccountsResp, err error) {
	// Call RPC service to get user OAuth accounts
	result, err := l.svcCtx.CoreRpc.GetUserOauthAccounts(l.ctx, &coreclient.GetUserOauthAccountsReq{
		UserId:   req.UserId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})

	if err != nil {
		return nil, err
	}

	// Convert RPC response to API response
	var accountInfos []types.OauthAccountInfo
	for _, account := range result.Data {
		accountInfo := types.OauthAccountInfo{
			BaseIDInfo: types.BaseIDInfo{
				Id:        account.Id,
				CreatedAt: account.CreatedAt,
				UpdatedAt: account.UpdatedAt,
			},
			UserId:           account.UserId,
			ProviderId:       account.ProviderId,
			ProviderType:     account.ProviderType,
			ProviderUserId:   account.ProviderUserId,
			ProviderUsername: account.ProviderUsername,
			ProviderNickname: account.ProviderNickname,
			ProviderEmail:    account.ProviderEmail,
			ProviderAvatar:   account.ProviderAvatar,
			TokenExpiresAt:   account.TokenExpiresAt,
			ExtraData:        account.ExtraData,
			LastLoginAt:      account.LastLoginAt,
			LastLoginIp:      account.LastLoginIp,
			LoginCount:       account.LoginCount,
		}
		accountInfos = append(accountInfos, accountInfo)
	}

	return &types.GetUserOauthAccountsResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "获取成功",
		},
		Data: types.OauthAccountListInfo{
			BaseListInfo: types.BaseListInfo{
				Total: result.Total,
			},
			Data: accountInfos,
		},
	}, nil
}
