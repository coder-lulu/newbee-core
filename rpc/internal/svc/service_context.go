package svc

import (
	"github.com/coder-lulu/newbee-common/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/ent"
	"github.com/coder-lulu/newbee-core/rpc/ent/hook"
	"github.com/coder-lulu/newbee-core/rpc/internal/config"
	"github.com/redis/go-redis/v9"

	"github.com/zeromicro/go-zero/core/logx"

	_ "github.com/coder-lulu/newbee-core/rpc/ent/runtime"
)

type ServiceContext struct {
	Config config.Config
	DB     *ent.Client
	Redis  redis.UniversalClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := ent.NewClient(
		ent.Log(logx.Error), // logger
		ent.Driver(c.DatabaseConf.NewNoCacheDriver()),
	)

	// add tenant hook (使用公共库)
	db.Use(hooks.TenantMutationHook())
	db.Intercept(hooks.TenantQueryInterceptor())

	// add data permission interceptor (使用公共库的统一Hook注册机制)
	// 注册需要数据权限控制的表
	hooks.RegisterDataPermissionInterceptorsWithTenant(db, "users", "departments", "positions", "roles")

	// add soft delete interceptor
	db.Intercept(hook.SoftDeleteInterceptor())

	rds := c.RedisConf.MustNewUniversalRedis()

	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  rds,
	}
}
