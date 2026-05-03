// Package listen contains logic for starting a http server
package listen

import (
	"fmt"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/service"
	"github.com/killi1812/go-cache-server/util/proc"
	"github.com/spf13/cobra"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

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

type listenParams struct {
	dig.In
	MgmtApi   app.CreateGinApi `name:"management"`
	DeployApi app.CreateGinApi `name:"deploy"`
	GC        *service.GCSrv
}

func listen(cmd *cobra.Command, args []string) error {
	foreground, err := cmd.Flags().GetBool(_FOREGROUND_FLAG_NAME)
	if err != nil {
		zap.S().DPanicf("Failed to retrieve foreground flag, err: %v", err)
	}

	if foreground {
		app.MultiStart(createApis())
	} else {
		err := proc.StartProcBackground(app.PID_FILE_NAME)
		if err != nil {
			return err
		}
		fmt.Printf("Server Started on management port %d and deploy port %d\n",
			config.Config.CacheServer.ServerPort, config.Config.CacheServer.DeployPort)
	}
	return nil
}

func createApis() map[string]app.CreateGinApi {
	mgmtAddr := fmt.Sprintf("%s:%d", config.Config.CacheServer.Hostname, config.Config.CacheServer.ServerPort)
	deployAddr := fmt.Sprintf("%s:%d", config.Config.CacheServer.Hostname, config.Config.CacheServer.DeployPort)

	zap.S().Debugf("Mapping Management API to %s", mgmtAddr)
	zap.S().Debugf("Mapping Deploy API to %s", deployAddr)

	var apis map[string]app.CreateGinApi

	app.Invoke(func(p listenParams) {
		p.GC.Start()

		apis = map[string]app.CreateGinApi{
			mgmtAddr:   p.MgmtApi,
			deployAddr: p.DeployApi,
		}
		zap.S().Debug("Successfully invoked named APIs and started background workers")
	})

	return apis
}
