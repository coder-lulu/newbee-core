package svc

import (
	"context"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/coder-lulu/newbee-common/config"
	"github.com/coder-lulu/newbee-common/i18n"
	"github.com/coder-lulu/newbee-common/middleware"
	"github.com/coder-lulu/newbee-common/utils/captcha"
	apiConfig "github.com/coder-lulu/newbee-core/api/internal/config"
	i18n2 "github.com/coder-lulu/newbee-core/api/internal/i18n"
	localMiddleware "github.com/coder-lulu/newbee-core/api/internal/middleware"
	"github.com/coder-lulu/newbee-core/rpc/coreclient"
	"github.com/coder-lulu/newbee-core/rpc/types/core"
	"github.com/suyuan32/simple-admin-job/jobclient"
	"github.com/suyuan32/simple-admin-message-center/mcmsclient"

	"github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// CoreRPCAdapter 适配器，将生成的gRPC客户端适配为中间件接口
type CoreRPCAdapter struct {
	client coreclient.Core
}

func (a *CoreRPCAdapter) InitRoleDataPermToRedis(ctx context.Context, req interface{}) (interface{}, error) {
	return a.client.InitRoleDataPermToRedis(ctx, &core.Empty{})
}

func (a *CoreRPCAdapter) InitDeptDataPermToRedis(ctx context.Context, req interface{}) (interface{}, error) {
	return a.client.InitDeptDataPermToRedis(ctx, &core.Empty{})
}

type ServiceContext struct {
	Config    apiConfig.Config
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

func NewServiceContext(c apiConfig.Config) *ServiceContext {
	rds := c.RedisConf.MustNewUniversalRedis()

	cbn := c.CasbinConf.MustNewCasbinWithOriginalRedisWatcher(c.DatabaseConf.Type, c.DatabaseConf.GetDSN(),
		c.RedisConf)

	trans := i18n.NewTranslator(c.I18nConf, i18n2.LocaleFS)

	coreClient := coreclient.NewCore(zrpc.NewClientIfEnable(c.CoreRpc))
	coreAdapter := &CoreRPCAdapter{client: coreClient}

	// 创建数据权限配置 - 默认启用租户模式
	dataPermConfig := &config.DataPermissionConfig{
		TenantMode: &config.TenantModeConfig{
			Enable:          true,  // 启用租户模式
			DefaultTenantId: 1,     // 默认租户ID
		},
		Cache: &config.DataPermCacheConfig{
			Expiration: 0, // 永不过期
		},
	}

	// 数据权限中间件配置
	middlewareConfig := &middleware.DataPermConfig{
		EnableTenantMode: dataPermConfig.TenantMode.Enable,
		DefaultTenantId:  dataPermConfig.TenantMode.DefaultTenantId,
		CacheExpiration:  dataPermConfig.Cache.Expiration,
	}

	return &ServiceContext{
		Config:    c,
		CoreRpc:   coreClient,
		JobRpc:    jobclient.NewJob(zrpc.NewClientIfEnable(c.JobRpc)),
		McmsRpc:   mcmsclient.NewMcms(zrpc.NewClientIfEnable(c.McmsRpc)),
		Captcha:   captcha.MustNewOriginalRedisCaptcha(c.Captcha, rds),
		Redis:     rds,
		Casbin:    cbn,
		Trans:     trans,
		Authority: localMiddleware.NewAuthorityMiddleware(cbn, rds).Handle,
		DataPerm:  middleware.NewDataPermMiddleware(rds, coreAdapter, trans, middlewareConfig).Handle,
	}
}
