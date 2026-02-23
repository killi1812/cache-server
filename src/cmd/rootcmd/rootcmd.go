// Package rootcmd exposes a root command of the program
package rootcmd

import (
	"fmt"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/util/db"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	verbose     = false
	doMigration = false
)

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use: "cache-server",
		// TODO: change descriptions
		Short:            "MyTool is a lightning fast CLI",
		Long:             "An example application to demonstrate Cobra's subcommand power.",
		Version:          app.Version,
		PersistentPreRun: setup,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	ptr.SetVersionTemplate(versionTempleta)

	ptr.PersistentFlags().StringVarP(&config.ConfigPath, "config", "c", "cache-server.conf", "path to config file")
	ptr.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	ptr.PersistentFlags().BoolVarP(&doMigration, "migration", "m", false, "preform auto migration")

	// remove cache-server help, only leave cache-server -h and cache-server --help
	// BUG: still in the compleating
	ptr.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
		Annotations: map[string]string{
			cobra.BashCompCustom: "__cache-server_no_suggestions",
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	})

	return ptr
}

// setup sets verbose logger and loads config in a PersistentPreRun
func setup(cmd *cobra.Command, args []string) {
	if verbose && app.Build == app.BuildProd {
		app.VerboseLoggerSetup()
	}
	zap.S().Debugf("Config file %s", config.ConfigPath)
	config.LoadConfig()

	if app.Build == app.BuildDev || doMigration {
		err := db.Migration(db.New())
		if err != nil {
			zap.S().DPanicf("Failed to run auto migration, err: %v", err)
			fmt.Println("Migration Failed")
		}
	}
}

var versionTempleta = `{{with .Name}}{{printf "%s " .}}{{end}}
Version:     ` + app.Version + `
Build Type:  ` + app.Build + `
Commit Hash: ` + app.CommitHash + `
Build Time:  ` + app.BuildTimestamp + `
`
