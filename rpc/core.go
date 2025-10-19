package main

import (
	"flag"
	"fmt"

	"github.com/coder-lulu/newbee-common/v2/orm/ent/hooks"
	"github.com/coder-lulu/newbee-core/rpc/internal/config"
	"github.com/coder-lulu/newbee-core/rpc/internal/server"
	"github.com/coder-lulu/newbee-core/rpc/internal/svc"
	"github.com/coder-lulu/newbee-core/rpc/types/core"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/core.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		core.RegisterCoreServer(grpcServer, server.NewCoreServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// ✅ 添加上下文传播拦截器，从gRPC incoming metadata中提取租户ID等信息并注入到context
	// 这确保了RPC服务端的logic能够通过context获取租户信息，统一Hook系统才能正确工作
	s.AddUnaryInterceptors(hooks.ContextPropagationServerInterceptor())
	s.AddStreamInterceptors(hooks.ContextPropagationStreamServerInterceptor())

	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
