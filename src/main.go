package main

import (
	"context"
	"os"

	"github.com/killi1812/go-cache-server/app"
	// "github.com/killi1812/go-cache-server/cmd/version"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var root *cobra.Command

func init() {
	app.Setup()

	root = &cobra.Command{
		Use:     "cache-server",
		Short:   "MyTool is a lightning fast CLI",
		Long:    `An example application to demonstrate Cobra's subcommand power.`,
		Version: app.Version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	root.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}
Version:     ` + app.Version + `
Build Type:  ` + app.Build + `
Commit Hash: ` + app.CommitHash + `
Build Time:  ` + app.BuildTimestamp + `
`)
}

func main() {
	ctx := context.Background()
	root.SetContext(ctx)
	if err := root.Execute(); err != nil {
		zap.S().Errorln(os.Stderr, err)
		os.Exit(1)
	}
}
