package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-common/orm/ent/entctx/tenantctx"
	"github.com/coder-lulu/newbee-core/api/internal/svc"
	"github.com/coder-lulu/newbee-core/api/internal/types"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserOauthProvidersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserOauthProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserOauthProvidersLogic {
	return &GetUserOauthProvidersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserOauthProvidersLogic) GetUserOauthProviders(req *types.UserOauthProviderListReq) (resp *types.UserOauthProviderListResp, err error) {
	// Get OAuth provider list from RPC
	enabledOnly := true
	if req.EnabledOnly != nil {
		enabledOnly = *req.EnabledOnly
	}

	// Build filter for enabled providers
	var providerListReq *coreclient.OauthProviderListReq
	if enabledOnly {
		providerListReq = &coreclient.OauthProviderListReq{
			Page:     1,
			PageSize: 100, // Get all enabled providers
		}
	} else {
		providerListReq = &coreclient.OauthProviderListReq{
			Page:     1,
			PageSize: 100,
		}
	}
	adminCtx := tenantctx.PublicCtx(l.ctx)

	providerResult, err := l.svcCtx.CoreRpc.GetOauthProviderList(adminCtx, providerListReq)
	if err != nil {
		return nil, err
	}

	// Convert to user provider info format
	var userProviders []types.UserOauthProviderInfo
	for _, provider := range providerResult.Data {
		// Skip disabled providers if enabledOnly is true
		if enabledOnly && (provider.Enabled == nil || !*provider.Enabled) {
			continue
		}

		userProvider := types.UserOauthProviderInfo{
			Id:          *provider.Id,
			Name:        *provider.Name,
			DisplayName: getDisplayName(provider),
			Type:        getProviderType(provider),
			IconUrl:     provider.IconUrl,
			IsBound:     false, // Default to false, will be updated below
		}

		userProviders = append(userProviders, userProvider)
	}

	return &types.UserOauthProviderListResp{
		BaseDataInfo: types.BaseDataInfo{
			Code: 0,
			Msg:  "获取成功",
		},
		Data: userProviders,
	}, nil
}

// getDisplayName returns the display name for the provider
func getDisplayName(provider *coreclient.OauthProviderInfo) string {
	if provider.DisplayName != nil && *provider.DisplayName != "" {
		return *provider.DisplayName
	}
	if provider.Name != nil {
		return *provider.Name
	}
	return "Unknown Provider"
}

// getProviderType returns the provider type
func getProviderType(provider *coreclient.OauthProviderInfo) string {
	if provider.Type != nil && *provider.Type != "" {
		return *provider.Type
	}
	if provider.ProviderType != nil && *provider.ProviderType != "" {
		return *provider.ProviderType
	}
	return "oauth"
}
