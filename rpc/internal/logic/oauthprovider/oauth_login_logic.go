package oauthprovider

import (
	"context"
	"strings"

	"golang.org/x/oauth2"

	"github.com/coder-lulu/newbee-core/rpc/ent/oauthprovider"

	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/internal/utils/dberrorhandler"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/logx"
)

var providerConfig = make(map[string]oauth2.Config)

// userInfoURL used to store infoURL in database | 用来存储获取用户信息网址数据
var userInfoURL = make(map[string]string)

type OauthLoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOauthLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OauthLoginLogic {
	return &OauthLoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OauthLoginLogic) OauthLogin(in *core.OauthLoginReq) (*core.OauthRedirectResp, error) {
	p, err := l.svcCtx.DB.OauthProvider.Query().Where(oauthprovider.NameEQ(in.Provider)).First(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, in)
	}

	// 🔐 解密client_secret用于OAuth流程
	var clientSecret string
	if p.EncryptedSecret != "" && p.EncryptionKeyID != "" {
		decrypted, err := l.svcCtx.EncryptionService.DecryptProviderSecret(p.EncryptedSecret, p.EncryptionKeyID)
		if err != nil {
			l.Logger.Errorw("Failed to decrypt client secret for OAuth", logx.Field("error", err), logx.Field("provider", p.Name))
			return nil, err
		}
		clientSecret = decrypted
	} else {
		// 兼容旧数据(未加密)
		clientSecret = p.ClientSecret
	}

	var config oauth2.Config
	if v, ok := providerConfig[p.Name]; ok {
		config = v
	} else {
		providerConfig[p.Name] = oauth2.Config{
			ClientID:     p.ClientID,
			ClientSecret: clientSecret, // ✅ 使用解密后的密钥
			Endpoint: oauth2.Endpoint{
				AuthURL:   replaceKeywords(p.AuthURL, p),
				TokenURL:  p.TokenURL,
				AuthStyle: oauth2.AuthStyle(p.AuthStyle),
			},
			RedirectURL: p.RedirectURL,
			Scopes:      strings.Split(p.Scopes, " "),
		}
		config = providerConfig[p.Name]
	}

	if _, ok := userInfoURL[p.Name]; !ok {
		userInfoURL[p.Name] = p.InfoURL
	}

	url := config.AuthCodeURL(in.State)

	return &core.OauthRedirectResp{Url: url}, nil
}
