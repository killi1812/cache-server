package main

import (
	"context"
	"os"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/cmd/rootcmd"
	"github.com/killi1812/go-cache-server/cmd/start"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rcmd *cobra.Command

func init() {
	app.Setup()

	rcmd = rootcmd.NewCmd()
	rcmd.AddCommand(start.NewCmd())
}

func main() {
	ctx := context.Background()

	if err := rcmd.ExecuteContext(ctx); err != nil {
		zap.S().Errorln(os.Stderr, err)
		os.Exit(1)
	}
}
