// Package listen contains logic for starting a http server
package listen

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	foreground = false
	attach     = false
)

// restart the app in foreground
var startcmd = exec.Command(os.Args[0], append(os.Args[1:], "--foreground")...)

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "listen",
		Short: "Start cache server",
		Long:  `Start cache server in the background`,
		Run:   listen,
	}

	ptr.PersistentFlags().BoolVarP(&foreground, "foreground", "f", false, "Run the app in foreground")
	return ptr
}

// TODO: check if it needed to be tread safe

func listen(cmd *cobra.Command, args []string) {
	if pid.CheckPid() {
		// return error app already running
		zap.S().Errorf("Error starting the server: %v", pid.ErrPidFileAlreadyExists)
		return
	}

	if foreground {
		// start the app foreground
		// TODO: check if .pid file should be created
		app.Start()
		// defer pid.RemovePid()
	} else {
		// start the app background
		zap.S().Debugf("Running command %s", startcmd.String())
		startcmd.Start()
		zap.S().Debugf("Running in background with PID: %d", startcmd.Process.Pid)

		err := pid.WritePid(startcmd.Process.Pid)
		if err != nil {
			zap.S().Errorf("Failed to write a pid to a file")
			zap.S().Errorf("Stopping the server")
			startcmd.Process.Signal(syscall.SIGTERM)
			startcmd.Wait()
			zap.S().Errorf("Server stopped")
		}
	}
}
