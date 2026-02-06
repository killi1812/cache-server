// Package start contains logic for starting a http server
package listen

import (
	"github.com/killi1812/go-cache-server/app"
	"github.com/spf13/cobra"
)

var foreground = false

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use: "listen",
		// TODO: change
		Short: "Start cache server",
		// TODO: add foreground option?
		Long: `Start cache server in the background`,
		Run:  listen,
	}

	ptr.PersistentFlags().BoolVarP(&foreground, "foreground", "f", false, "run the app in foreground")
	return ptr
}

func listen(cmd *cobra.Command, args []string) {
	if foreground {
		app.Start()
	}
}
