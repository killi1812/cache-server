// Package rootcmd exposes a root command of the program
package rootcmd

import (
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	ptr := &cobra.Command{
		Use: "cache-server",
		// TODO: change
		Short:   "MyTool is a lightning fast CLI",
		Long:    `An example application to demonstrate Cobra's subcommand power.`,
		Version: app.Version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	ptr.SetVersionTemplate(versionTempleta)
	ptr.PersistentFlags().StringVarP(&config.ConfigPath, "config", "c", "cache-server.conf", "path to config file")
	return ptr
}

var versionTempleta = `{{with .Name}}{{printf "%s " .}}{{end}}
Version:     ` + app.Version + `
Build Type:  ` + app.Build + `
Commit Hash: ` + app.CommitHash + `
Build Time:  ` + app.BuildTimestamp + `
`
