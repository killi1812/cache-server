package main

import (
	"context"
	"os"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/cmd/agent"
	"github.com/killi1812/go-cache-server/cmd/cache"
	"github.com/killi1812/go-cache-server/cmd/listen"
	"github.com/killi1812/go-cache-server/cmd/rootcmd"
	"github.com/killi1812/go-cache-server/cmd/stop"
	"github.com/killi1812/go-cache-server/cmd/storepath"
	"github.com/killi1812/go-cache-server/cmd/workspace"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/spf13/cobra"
)

var rcmd *cobra.Command

func init() {
	app.Setup()

	// registering commands
	rcmd = rootcmd.NewCmd()
	rcmd.AddCommand(listen.NewCmd())
	rcmd.AddCommand(stop.NewCmd())
	rcmd.AddCommand(cache.NewCmd())
	rcmd.AddCommand(agent.NewCmd())
	rcmd.AddCommand(workspace.NewCmd())
	rcmd.AddCommand(storepath.NewCmd())

	// Provide storage options
	app.Provide(db.New)
	app.Provide(objstor.New)

	// Provide services
	app.Provide(service.NewAgentSrv)
	app.Provide(service.NewCacheSrv)
	app.Provide(service.NewStorePathSrv)
	app.Provide(service.NewWorkspaceSrv)
	app.Provide(service.NewDeploymentSrv)
}

//	@title			Cache Server API
//	@version		1.0
//	@description	This is a nix cache server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer <your-jwt-token>" to authenticate

func main() {
	ctx := context.Background()

	if err := rcmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
