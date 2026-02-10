// Package listen contains logic for starting a http server
package listen

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var foreground = false

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
	if foreground {
		// start the app foreground
		app.Start()
	} else {
		if pid.CheckPid() {
			// return error app already running
			zap.S().Errorf("Error starting the server: %v", pid.ErrPidFileAlreadyExists)
			return
		}

		// start the app in background
		absPath, _ := filepath.Abs(os.Args[0])
		args := strings.Join(append(os.Args[1:], "--foreground"), " ")

		zap.S().Debugf("Starting Process: %s %s", absPath, args)

		cmd := exec.Command(absPath, strings.Split(args, " ")...)
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}

		err := cmd.Start()
		process := cmd.Process
		if err != nil {
			zap.S().Fatalf("Failed to start: %v", err)
		}

		zap.S().Debugf("Running in background with PID: %d", process.Pid)

		err = pid.WritePid(process.Pid)
		if err != nil {
			zap.S().Errorf("Failed to write a pid to a file")
			zap.S().Errorf("Stopping the server")
			process.Signal(syscall.SIGTERM)
			process.Wait()
			zap.S().Errorf("Server stopped")
			return
		}

		// This tells Go "I am not going to Wait() for this, let it run"
		err = cmd.Process.Release()
		if err != nil {
			zap.S().Errorf("Failed to release process: %v", err)
			process.Signal(syscall.SIGTERM)
			process.Wait()
			zap.S().Errorf("Server stopped")
			return
		}
	}
}
