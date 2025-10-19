package config

import (
	"github.com/coder-lulu/newbee-common/v2/plugins/casbin"
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/coder-lulu/newbee-common/v2/config"
)

type Config struct {
	zrpc.RpcServerConf
	DatabaseConf config.DatabaseConf
	CasbinConf   casbin.CasbinConf
	RedisConf    config.RedisConf
	EncryptionKey string `json:",optional"` // OAuth Provider加密密钥
}
