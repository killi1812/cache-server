package main

import (
	"context"
	"os"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/cmd/rootcmd"
	"github.com/killi1812/go-cache-server/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	rcmd    *cobra.Command
	verbose bool
)

func init() {
	app.Setup()

	rcmd = rootcmd.NewRootCommand()
	rcmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "verbose output")
	rcmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if verbose && app.Build == app.BuildProd {
			app.VerboseLoggerSetup()
		}
		zap.S().Debugf("Config file %s", config.ConfigPath)
	}
}

func main() {
	ctx := context.Background()

	if err := rcmd.ExecuteContext(ctx); err != nil {
		zap.S().Errorln(os.Stderr, err)
		os.Exit(1)
	}
}
