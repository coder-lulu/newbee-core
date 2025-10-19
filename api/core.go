//	core
//
//	Description: core service
//
//	Schemes: http, https
//	Host: localhost:0
//	BasePath: /
//	Version: 0.0.1
//	SecurityDefinitions:
//	  Token:
//	    type: apiKey
//	    name: Authorization
//	    in: header
//	Security:
//	  Token:
//	Consumes:
//	  - application/json
//
//	Produces:
//	  - application/json
//
// swagger:meta
package main

import (
	"flag"
	"fmt"

	"github.com/coder-lulu/newbee-common/middleware/integration"
	"github.com/coder-lulu/newbee-core/api/internal/config"
	"github.com/coder-lulu/newbee-core/api/internal/handler"
	"github.com/coder-lulu/newbee-core/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/core.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf, rest.WithCors(c.CROSConf.Address))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer func() {
		if ctx.IntegrationResult != nil && ctx.IntegrationResult.Manager != nil {
			if err := ctx.IntegrationResult.Manager.Shutdown(); err != nil {
				logx.Errorf("failed to shutdown middleware manager: %v", err)
			}
		}
	}()

	// 🎉 使用统一的集成API应用中间件链
	integration.ApplyToServer(server, ctx.IntegrationResult)

	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
