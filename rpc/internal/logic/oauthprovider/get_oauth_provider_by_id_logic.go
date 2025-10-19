package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/typeconv"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthProviderByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOauthProviderByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthProviderByIdLogic {
	return &GetOauthProviderByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOauthProviderByIdLogic) GetOauthProviderById(in *core.IDReq) (*core.OauthProviderInfo, error) {
	result, err := l.svcCtx.DB.OauthProvider.Get(l.ctx, in.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 🔐 解密client_secret (仅用于管理接口展示)
	var clientSecret *string
	if result.EncryptedSecret != "" && result.EncryptionKeyID != "" {
		decrypted, err := l.svcCtx.EncryptionService.DecryptProviderSecret(result.EncryptedSecret, result.EncryptionKeyID)
		if err != nil {
			l.Logger.Errorw("Failed to decrypt client secret", logx.Field("error", err))
			// 解密失败时返回掩码值,不阻塞请求
			masked := "***SECRET_ENCRYPTED***"
			clientSecret = &masked
		} else {
			clientSecret = &decrypted
		}
	}

	return &core.OauthProviderInfo{
		Id:           &result.ID,
		CreatedAt:    pointy.GetPointer(result.CreatedAt.UnixMilli()),
		UpdatedAt:    pointy.GetPointer(result.UpdatedAt.UnixMilli()),
		Name:         &result.Name,
		ClientId:     &result.ClientID,
		ClientSecret: clientSecret, // ✅ 返回解密后的密钥(用于管理界面)
		RedirectUrl:  &result.RedirectURL,
		Scopes:       &result.Scopes,
		AuthUrl:      &result.AuthURL,
		TokenUrl:     &result.TokenURL,
		AuthStyle:    typeconv.ConvertAuthStyleFromEnt(result.AuthStyle),
		InfoUrl:      &result.InfoURL,
		// Enhanced fields from OAuth refactor
		DisplayName:     &result.DisplayName,
		Type:            &result.Type,
		ProviderType:    &result.ProviderType,
		EncryptedSecret: &result.EncryptedSecret,
		EncryptionKeyId: &result.EncryptionKeyID,
		ExtraConfig:     typeconv.ConvertExtraConfigFromEnt(result.ExtraConfig),
		Enabled:         &result.Enabled,
		Sort:            &result.Sort,
		Remark:          &result.Remark,
		SupportPkce:     &result.SupportPkce,
		IconUrl:         &result.IconURL,
		CacheTtl:        typeconv.ConvertCacheTTLFromEnt(result.CacheTTL),
		WebhookUrl:      &result.WebhookURL,
		SuccessCount:    typeconv.ConvertCountFromEnt(result.SuccessCount),
		FailureCount:    typeconv.ConvertCountFromEnt(result.FailureCount),
		LastUsedAt:      typeconv.ConvertLastUsedAtFromEnt(result.LastUsedAt),
		// Tenant and status fields
		Status:   typeconv.ConvertStatusFromEnt(result.Status),
		TenantId: &result.TenantID,
	}, nil
}
