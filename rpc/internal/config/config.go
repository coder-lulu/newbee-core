package config

import (
	"github.com/coder-lulu/newbee-common/plugins/casbin"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/coder-lulu/newbee-common/config"
)

type Config struct {
	zrpc.RpcServerConf
	DatabaseConf config.DatabaseConf
	CasbinConf   casbin.CasbinConf
	RedisConf    config.RedisConf
}
