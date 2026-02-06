// Package pid contains read and write operations for pid file
package pid

import (
	"errors"
	"os"
	"strconv"

	"go.uber.org/zap"
)

const _PID_FILE_NAME = "/tmp/cache-server.pid"

// TODO: write test

// WritePid will write a given pid to a /tmp/cache-server.pid file
func WritePid(pid int) error {
	content := []byte(strconv.Itoa(pid))

	// 0644 sets permissions: owner can read/write, others can only read
	f, err := os.OpenFile(_PID_FILE_NAME, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			zap.S().Errorf("Error The file '%s' already exists", _PID_FILE_NAME)
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

// StatPid check if .pid file exists
func StatPid() bool {
	return false
}

// RemovePid will remove .pid file
func RemovePid() error {
	return nil
}

// ReadPid will read the .pid file and return -1,error if it fails
func ReadPid() (int, error) {
	return -1, nil
}
