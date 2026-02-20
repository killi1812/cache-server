// Package listen contains logic for starting a http server
package listen

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/killi1812/go-cache-server/app"
	"github.com/killi1812/go-cache-server/config"
	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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

// TODO: check if it needed to be tread safe

func listen(cmd *cobra.Command, args []string) error {
	if foreground {
		// start the app foreground
		//
		addr := fmt.Sprintf("%s:%d", config.Config.CacheServer.Hostname, config.Config.CacheServer.ServerPort)
		app.Start(nil, addr)
	} else {
		err := startBackground()
		if err != nil {
			return ErrFailedToStart
		}

	}
	return nil
}

func startBackground() error {
	if pid.CheckPid(app.PID_FILE_NAME) {
		// return error app already running
		zap.S().Debugf("Error starting the server: %v", pid.ErrPidFileAlreadyExists)
		return pid.ErrPidFileAlreadyExists
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
		zap.S().Debugf("Failed to start: %v", err)
	}

	zap.S().Debugf("Running in background with PID: %d", process.Pid)

	err = pid.WritePid(app.PID_FILE_NAME, process.Pid)
	if err != nil {
		zap.S().Debugf("Failed to write a pid to a file")
		zap.S().Debugf("Stopping the server")
		process.Signal(syscall.SIGTERM)
		process.Wait()
		zap.S().Debugf("Server stopped")
		return err
	}

	// This tells Go "I am not going to Wait() for this, let it run"
	err = cmd.Process.Release()
	if err != nil {
		zap.S().Debugf("Failed to release process: %v", err)
		process.Signal(syscall.SIGTERM)
		process.Wait()
		zap.S().Debugf("Server stopped")
		return err
	}
	return nil
}
