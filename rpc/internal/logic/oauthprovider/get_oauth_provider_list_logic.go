package oauthprovider

import (
	"context"

	"github.com/coder-lulu/newbee-common/v2/utils/pointy"

	"github.com/coder-lulu/newbee-core/rpc/ent/oauthprovider"
	"github.com/coder-lulu/newbee-core/rpc/ent/predicate"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/typeconv"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOauthProviderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOauthProviderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOauthProviderListLogic {
	return &GetOauthProviderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOauthProviderListLogic) GetOauthProviderList(in *core.OauthProviderListReq) (*core.OauthProviderListResp, error) {
	var predicates []predicate.OauthProvider
	if in.Name != nil {
		predicates = append(predicates, oauthprovider.NameContains(*in.Name))
	}
	result, err := l.svcCtx.DB.OauthProvider.Query().Where(predicates...).Page(l.ctx, in.Page, in.PageSize)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	resp := &core.OauthProviderListResp{}
	resp.Total = result.PageDetails.Total

	for _, v := range result.List {
		// 🔐 解密client_secret (列表接口返回掩码值，不解密完整密钥)
		var clientSecret *string
		if v.EncryptedSecret != "" && v.EncryptionKeyID != "" {
			// 列表接口为安全考虑，只返回掩码
			masked := "******" // 6个星号表示已加密
			clientSecret = &masked
		} else if v.ClientSecret != "" {
			// 兼容旧数据：返回掩码
			masked := "******"
			clientSecret = &masked
		}
		
		resp.Data = append(resp.Data, &core.OauthProviderInfo{
			Id:           &v.ID,
			CreatedAt:    pointy.GetPointer(v.CreatedAt.UnixMilli()),
			UpdatedAt:    pointy.GetPointer(v.UpdatedAt.UnixMilli()),
			Name:         &v.Name,
			ClientId:     &v.ClientID,
			ClientSecret: clientSecret, // ✅ 列表接口返回掩码值
			RedirectUrl:  &v.RedirectURL,
			Scopes:       &v.Scopes,
			AuthUrl:      &v.AuthURL,
			TokenUrl:     &v.TokenURL,
			AuthStyle:    typeconv.ConvertAuthStyleFromEnt(v.AuthStyle),
			InfoUrl:      &v.InfoURL,
			// Enhanced fields from OAuth refactor
			DisplayName:     &v.DisplayName,
			Type:            &v.Type,
			ProviderType:    &v.ProviderType,
			EncryptedSecret: &v.EncryptedSecret,
			EncryptionKeyId: &v.EncryptionKeyID,
			ExtraConfig:     typeconv.ConvertExtraConfigFromEnt(v.ExtraConfig),
			Enabled:         &v.Enabled,
			Sort:            &v.Sort,
			Remark:          &v.Remark,
			SupportPkce:     &v.SupportPkce,
			IconUrl:         &v.IconURL,
			CacheTtl:        typeconv.ConvertCacheTTLFromEnt(v.CacheTTL),
			WebhookUrl:      &v.WebhookURL,
			SuccessCount:    typeconv.ConvertCountFromEnt(v.SuccessCount),
			FailureCount:    typeconv.ConvertCountFromEnt(v.FailureCount),
			LastUsedAt:      typeconv.ConvertLastUsedAtFromEnt(v.LastUsedAt),
			// Tenant and status fields
			Status:   typeconv.ConvertStatusFromEnt(v.Status),
			TenantId: &v.TenantID,
		})
	}

	return resp, nil
}
