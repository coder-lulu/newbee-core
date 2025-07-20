package svc

import (
	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/suyuan32/simple-admin-common/i18n"
	"github.com/suyuan32/simple-admin-common/utils/captcha"
	"github.com/coder-lulu/newbee-core/api/internal/config"
	i18n2 "github.com/coder-lulu/newbee-core/api/internal/i18n"
	"github.com/coder-lulu/newbee-core/api/internal/middleware"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/suyuan32/simple-admin-job/jobclient"
	"github.com/suyuan32/simple-admin-message-center/mcmsclient"

	"github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	Authority rest.Middleware
	DataPerm  rest.Middleware
	CoreRpc   coreclient.Core
	JobRpc    jobclient.Job
	McmsRpc   mcmsclient.Mcms
	Redis     redis.UniversalClient
	Casbin    *casbin.Enforcer
	Trans     *i18n.Translator
	Captcha   *base64Captcha.Captcha
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds := c.RedisConf.MustNewUniversalRedis()

	cbn := c.CasbinConf.MustNewCasbinWithOriginalRedisWatcher(c.DatabaseConf.Type, c.DatabaseConf.GetDSN(),
		c.RedisConf)

	trans := i18n.NewTranslator(c.I18nConf, i18n2.LocaleFS)

	coreClient := coreclient.NewCore(zrpc.NewClientIfEnable(c.CoreRpc))

	return &ServiceContext{
		Config:    c,
		CoreRpc:   coreClient,
		JobRpc:    jobclient.NewJob(zrpc.NewClientIfEnable(c.JobRpc)),
		McmsRpc:   mcmsclient.NewMcms(zrpc.NewClientIfEnable(c.McmsRpc)),
		Captcha:   captcha.MustNewOriginalRedisCaptcha(c.Captcha, rds),
		Redis:     rds,
		Casbin:    cbn,
		Trans:     trans,
		Authority: middleware.NewAuthorityMiddleware(cbn, rds).Handle,
		DataPerm:  middleware.NewDataPermMiddleware(rds, coreClient, trans).Handle,
	}
}
