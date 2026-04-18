// Package listen contains logic for starting a http server
package listen

import (
	"errors"
	"fmt"

	"github.com/killi1812/go-cache-server/api"
	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/objstor"
	"github.com/killi1812/go-cache-server/util/proc"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ErrFailedToStart = errors.New("failed to start the server")

const _FOREGROUND_FLAG_NAME = "foreground"

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "listen",
		Short: "Start cache server",
		Long:  `Start cache server in the background`,
		RunE:  listen,
	}

	ptr.PersistentFlags().BoolP(_FOREGROUND_FLAG_NAME, "f", false, "Run the app in foreground")
	return ptr
}

func listen(cmd *cobra.Command, args []string) error {
	foreground, err := cmd.Flags().GetBool(_FOREGROUND_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve foreground flag, err: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", config.Config.CacheServer.Hostname, config.Config.CacheServer.ServerPort)
	if foreground {
		// start the app foreground
		app.Invoke(func(
			cs *service.CacheSrv,
			ps *service.StorePathSrv,
			as *service.AgentSrv,
			ws *service.WorkspaceSrv,
			ds *service.DeploymentSrv,
			h *service.Hub,
			storage objstor.ObjectStorage,
		) {
			app.Start(api.NewApi(cs, ps, as, ws, ds, h, storage), addr)
		})
	} else {
		err := proc.StartProcBackground(app.PID_FILE_NAME)
		if err != nil {
			return err
		}
		fmt.Printf("Server Started:\t http://%s\n", addr)
	}
	return nil
}
