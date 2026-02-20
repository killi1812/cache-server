// Package listen contains logic for starting a http server
package listen

import (
	"errors"
	"fmt"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/util/proc"
	"github.com/spf13/cobra"
)

var (
	foreground       = false
	ErrFailedToStart = errors.New("failed to start the server")
)

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "listen",
		Short: "Start cache server",
		Long:  `Start cache server in the background`,
		RunE:  listen,
	}

	ptr.PersistentFlags().BoolVarP(&foreground, "foreground", "f", false, "Run the app in foreground")
	return ptr
}

func listen(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", config.Config.CacheServer.Hostname, config.Config.CacheServer.ServerPort)
	if foreground {
		// start the app foreground
		app.Start(nil, addr)
	} else {
		err := proc.StartProcBackground(app.PID_FILE_NAME)
		if err != nil {
			return err
		}
		fmt.Printf("Server Started:\t http://%s\n", addr)
	}
	return nil
}
