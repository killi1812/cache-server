package help

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cache-server",
	Short: "MyTool is a lightning fast CLI",
	Long:  `An example application to demonstrate Cobra's subcommand power.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
