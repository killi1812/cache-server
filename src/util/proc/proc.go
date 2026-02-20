// Package proc handles spawning processes
package proc

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/killi1812/go-cache-server/util/pid"
	"go.uber.org/zap"
)

var (
	ErrFailedToStart = errors.New("failed to start the server")
	ErrFailedToStop  = errors.New("failed stopping the server")
)

func StartProcBackground(pidFilename string) error {
	if pid.CheckPid(pidFilename) {
		// return error app already running
		zap.S().Debugf("Error starting the server: %v", pid.ErrPidFileAlreadyExists)
		return pid.ErrPidFileAlreadyExists
	}

	// start the app in background
	absPath, _ := filepath.Abs(os.Args[0])
	args := strings.Join(append(os.Args[1:], "--foreground"), " ")

	zap.S().Infof("Starting Process: %s %s", absPath, args)

	cmd := exec.Command(absPath, strings.Split(args, " ")...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	err := cmd.Start()
	if err != nil {
		zap.S().Errorf("Failed to start: %v", err)
	}

	zap.S().Infof("Running in background with PID: %d", cmd.Process.Pid)

	err = pid.WritePid(pidFilename, cmd.Process.Pid)
	if err != nil {
		zap.S().Errorf("Failed to write a pid to a file")
		zap.S().Infof("Stopping the server")
		cmd.Process.Signal(syscall.SIGTERM)
		cmd.Process.Wait()
		zap.S().Infof("Server stopped")
		return errors.Join(ErrFailedToStart, err)
	}

	// This tells Go "I am not going to Wait() for this, let it run"
	err = cmd.Process.Release()
	if err != nil {
		zap.S().Errorf("Failed to release process: %v", err)
		cmd.Process.Signal(syscall.SIGTERM)
		cmd.Process.Wait()
		zap.S().Infof("Server stopped")
		return errors.Join(ErrFailedToStart, err)
	}
	return nil
}

type StopOpts struct {
	NoAsk  bool // don't ask for user confirmation use defaults
	Search bool // search for proc just fail
}

func StopProc(filename string, opts ...StopOpts) error {
	var opt StopOpts
	if len(opts) == 1 {
		opt = opts[0]
	}

	procPid := -1

	// find pid
	if !pid.CheckPid(filename) {
		zap.S().Warn("No pid file")
		if !opt.Search {
			return ErrFailedToStop
		}

		p, err := findPid(filename, opt.NoAsk)
		if err != nil {
			return errors.Join(ErrFailedToStop, err)
		}
		procPid = p
	} else {
		p, err := readPid(filename)
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

	return nil
}

func IsRunning(name string) bool {
	return pid.CheckPid(name)

	// if pid.CheckPid(name) {
	// 	return true
	// }
	// _, err := pid.FindPidByName(name)
	// return err == nil
}

// findPid trys to find program pid, ask for confirmation if it is found
func findPid(name string, noAsk bool) (int, error) {
	zap.S().Warn("Trying to find the process by name")
	// check for cache-server process and ask to stop it
	p, err := pid.FindPidByName(name)
	if err != nil {
		zap.S().Errorf("No process with name '%s' is running, err: %v", name, err)
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

func readPid(filename string) (int, error) {
	p, err := pid.ReadPid(filename)
	if err != nil {
		zap.S().Errorf("Failed to read pid file, err: %+v", err)
		return -1, err
	}
	defer pid.RemovePid(filename)

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
