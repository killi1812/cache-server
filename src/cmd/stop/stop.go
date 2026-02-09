// Package stop
package stop

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var noAsk = false

func NewCmd() *cobra.Command {
	ptr := &cobra.Command{
		Use:   "stop",
		Short: "Stop cache server",
		Long:  `Stop cache server running in the background`,
		RunE:  stop,
	}

	ptr.PersistentFlags().BoolVarP(&noAsk, "no-ask", "n", false, "don't ask questions assume default answer for all")

	return ptr
}

func stop(cmd *cobra.Command, args []string) error {
	procPid := -1
	// check for .pid file
	if !pid.CheckPid() {
		zap.S().Errorf("No pid file")

		// check for cache-server process and ask to stop it
		p, err := pid.FindPidByName()
		if err != nil {
			zap.S().Debug("No process with name cache-server is running")
			zap.S().Error("Failed stopping the cache-server")
			return err
		}
		zap.S().Infof("Found a cache-server proceess with pid %d ", p)
		if !noAsk {
			fmt.Print("do you want to stop it? [Y/n]: ")
			scanner := bufio.NewScanner(os.Stdin)

			if scanner.Scan() {
				response := strings.ToLower(strings.TrimSpace(scanner.Text()))
				if response == "n" || response == "no" {
					zap.S().Info("Not stopping the cache-server")
					return nil
				}
			}
		}
		// set pid to found process pid
		procPid = p
	} else {
		p, err := pid.ReadPid()
		if err != nil {
			zap.S().Debugf("Failed to read pid file, err: %+v", err)
			zap.S().Error("Failed stopping the cache-server")
			return err
		}
		defer pid.RemovePid()
		// set pid to .pid file value
		procPid = p
	}

	proc, err := os.FindProcess(procPid)
	if err != nil {
		zap.S().Debugf("Failed to find the process with pid %d, err: %+v", procPid, err)
		zap.S().Error("Failed stopping the cache-server")
		return err
	}

	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		zap.S().Debugf("Failed to send SIGTERM signal to the process, err: %+v", err)
		zap.S().Error("Failed stopping the cache-server")
		return err
	}

	return nil
}
