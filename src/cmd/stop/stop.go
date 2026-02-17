// Package stop
package stop

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var noAsk = false

var ErrFailedToStop = errors.New("failed stopping the cache-server")

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

	// find pid
	if !pid.CheckPid() {
		zap.S().Warn("No pid file")
		p, err := findPid()
		if err != nil {
			return ErrFailedToStop
		}
		procPid = p
	} else {
		p, err := readPid()
		if err != nil {
			return ErrFailedToStop
		}
		procPid = p
	}

	if procPid == -1 {
		return nil
	}

	err := stopProc(procPid)
	if err != nil {
		return ErrFailedToStop
	}

	fmt.Println("Server stopped successfully")
	return nil
}

// findPid trys to find program pid, ask for confirmation if it is found
func findPid() (int, error) {
	zap.S().Warn("Trying to find the process by name")
	// check for cache-server process and ask to stop it
	p, err := pid.FindPidByName()
	if err != nil {
		zap.S().Errorf("No process with name cache-server is running")
		zap.S().Errorf("Error: %v", err)
		return -1, err
	}
	zap.S().Infof("Found a cache-server process with pid %d ", p)
	fmt.Printf("Found a cache-server process with pid %d\n", p)

	if !noAsk {
		fmt.Print("Do you want to stop it? [Y/n]: ")
		scanner := bufio.NewScanner(os.Stdin)

		if scanner.Scan() {
			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if response == "n" || response == "no" {
				zap.S().Infof("user choice (%d) Not stopping the cache-server", response)
				fmt.Printf("Not stopping the cache-server")
				return -1, nil
			}
		}
	}

	return p, nil
}

func readPid() (int, error) {
	p, err := pid.ReadPid()
	if err != nil {
		zap.S().Errorf("Failed to read pid file, err: %+v", err)
		return -1, err
	}
	defer pid.RemovePid()

	return p, nil
}

func stopProc(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		zap.S().Errorf("Failed to find the process with pid %d, err: %+v", pid, err)
		return err
	}

	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		zap.S().Errorf("Failed to send SIGTERM signal to the process, err: %+v", err)
		return err
	}
	return nil
}
