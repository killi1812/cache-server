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
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rcmd *cobra.Command

func init() {
	app.Setup()

	rcmd = rootcmd.NewCmd()
	rcmd.AddCommand(listen.NewCmd())
	rcmd.AddCommand(stop.NewCmd())
	rcmd.AddCommand(cache.NewCmd())
	rcmd.AddCommand(agent.NewCmd())
	rcmd.AddCommand(workspace.NewCmd())
	rcmd.AddCommand(storepath.NewCmd())
}

func main() {
	ctx := context.Background()

	if err := rcmd.ExecuteContext(ctx); err != nil {
		zap.S().Errorln(os.Stderr, err)
		os.Exit(1)
	}
}
