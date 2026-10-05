package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/i18n"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/typeconv"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOauthProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOauthProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOauthProviderLogic {
	return &CreateOauthProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateOauthProviderLogic) CreateOauthProvider(in *core.OauthProviderInfo) (*core.BaseIDResp, error) {
	// 🔐 加密client_secret
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

	result, err := l.svcCtx.DB.OauthProvider.Create().
		SetNotNilName(in.Name).
		SetNotNilClientID(in.ClientId).
		// The legacy column is required; new credentials live only in encrypted_secret.
		SetClientSecret("").
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
		SetNotNilEncryptedSecret(encryptedSecret). // ✅ 存储加密后的密钥
		SetNotNilEncryptionKeyID(encryptionKeyID). // ✅ 存储密钥ID
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
		// Tenant and status fields - use standard ent methods
		SetNillableStatus(typeconv.ConvertStatus(in.Status)).
		SetNillableTenantID(in.TenantId).
		Save(l.ctx)

	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	return &core.BaseIDResp{Id: result.ID, Msg: i18n.CreateSuccess}, nil
}
