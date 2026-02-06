// Package pid contains read and write operations for pid file
package pid

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

const _PID_FILE_NAME = "/tmp/cache-server.pid"

var ErrPidFileAlreadyExists = fmt.Errorf("error the file '%s' already exists", _PID_FILE_NAME)

// TODO: write test

// WritePid will write a given pid to a /tmp/cache-server.pid file
func WritePid(pid int) error {
	content := []byte(strconv.Itoa(pid))

	// 0644 sets permissions: owner can read/write, others can only read
	f, err := os.OpenFile(_PID_FILE_NAME, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			zap.S().Error(ErrPidFileAlreadyExists)
		} else {
			zap.S().Errorf("Error opening file:", err)
		}
		return err
	}
	defer f.Close()

	// Write the content
	_, err = f.Write(content)
	if err != nil {
		zap.S().Errorf("Error writing a Pid to file = %s, err = %v", _PID_FILE_NAME, err)
		return err
	}

	zap.S().Debugf("Successfully wrote pid to %s\n", _PID_FILE_NAME)

	return nil
}

// CheckPid check if .pid file exists
func CheckPid() bool {
	_, err := os.ReadFile(_PID_FILE_NAME)
	return err == nil
}

// RemovePid will remove .pid file
func RemovePid() error {
	zap.S().Debugf("Removeing the pid file %s", _PID_FILE_NAME)
	return os.Remove(_PID_FILE_NAME)
}

// ReadPid will read the .pid file and return -1,error if it fails
func ReadPid() (int, error) {
	// 1. Read the PID file
	data, err := os.ReadFile(_PID_FILE_NAME)
	if err != nil {
		return -1, err
	}

	// 2. Parse the PID
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return -1, err
	}

	return pid, nil
}
