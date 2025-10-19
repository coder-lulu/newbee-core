package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/typeconv"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/coder-lulu/newbee-common/v2/i18n"
)

type UpdateOauthProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateOauthProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOauthProviderLogic {
	return &UpdateOauthProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateOauthProviderLogic) UpdateOauthProvider(in *core.OauthProviderInfo) (*core.BaseResp, error) {
	// 🔐 加密client_secret (仅当提供新密钥时)
	var encryptedSecret *string
	var encryptionKeyID *string
	
	if in.ClientSecret != nil && *in.ClientSecret != "" {
		encrypted, keyID, err := l.svcCtx.EncryptionService.EncryptProviderSecret(*in.ClientSecret)
		if err != nil {
			l.Logger.Errorw("Failed to encrypt client secret", logx.Field("error", err))
			return nil, err
		}
		encryptedSecret = &encrypted
		encryptionKeyID = &keyID
		
		// 清除明文密钥
		in.ClientSecret = nil
	}
	
	err := l.svcCtx.DB.OauthProvider.UpdateOneID(*in.Id).
		SetNotNilName(in.Name).
		SetNotNilClientID(in.ClientId).
		// SetNotNilClientSecret(in.ClientSecret). // ❌ 不再更新明文
		SetNotNilRedirectURL(in.RedirectUrl).
		SetNotNilScopes(in.Scopes).
		SetNotNilAuthURL(in.AuthUrl).
		SetNotNilTokenURL(in.TokenUrl).
		SetNotNilAuthStyle(typeconv.ConvertAuthStyle(in.AuthStyle)).
		SetNotNilInfoURL(in.InfoUrl).
		// Enhanced fields from OAuth refactor
		SetNotNilDisplayName(in.DisplayName).
		SetNotNilType(in.Type).
		SetNotNilProviderType(in.ProviderType).
		SetNotNilEncryptedSecret(encryptedSecret). // ✅ 更新加密密钥(如果提供)
		SetNotNilEncryptionKeyID(encryptionKeyID). // ✅ 更新密钥ID(如果提供)
		SetNotNilExtraConfig(typeconv.ConvertExtraConfig(in.ExtraConfig)).
		SetNotNilEnabled(in.Enabled).
		SetNotNilSort(in.Sort).
		SetNotNilRemark(in.Remark).
		SetNotNilSupportPkce(in.SupportPkce).
		SetNotNilIconURL(in.IconUrl).
		SetNotNilCacheTTL(typeconv.ConvertCacheTTL(in.CacheTtl)).
		SetNotNilWebhookURL(in.WebhookUrl).
		SetNotNilSuccessCount(typeconv.ConvertCount(in.SuccessCount)).
		SetNotNilFailureCount(typeconv.ConvertCount(in.FailureCount)).
		SetNotNilLastUsedAt(typeconv.ConvertLastUsedAt(in.LastUsedAt)).
		// Status field - use standard ent methods
		SetNillableStatus(typeconv.ConvertStatus(in.Status)).
		Exec(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	if _, ok := providerConfig[*in.Name]; ok {
		delete(providerConfig, *in.Name)
	}

	return &core.BaseResp{Msg: i18n.UpdateSuccess}, nil
}
