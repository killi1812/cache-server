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

// WritePid will write a given pid to a /tmp/cache-server.pid file
func WritePid(pid int) error { return writePid(pid, _PID_FILE_NAME) }

// CheckPid check if .pid file exists
func CheckPid() bool { return checkPid(_PID_FILE_NAME) }

// RemovePid will remove .pid file
func RemovePid() error { return removePid(_PID_FILE_NAME) }

// ReadPid will read the .pid file and return -1,error if it fails
func ReadPid() (int, error) { return readPid(_PID_FILE_NAME) }

func writePid(pid int, filepath string) error {
	content := []byte(strconv.Itoa(pid))

	// 0444 sets permissions: file is readonly
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o444)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			zap.S().Error(ErrPidFileAlreadyExists)
		} else {
			zap.S().Errorf("Error opening file:", err)
		}
		return err
	}
	defer f.Close()

	_, err = f.Write(content)
	if err != nil {
		zap.S().Errorf("Error writing a Pid to file = %s, err = %v", filepath, err)
		return err
	}

	zap.S().Debugf("Successfully wrote pid to %s\n", filepath)

	return nil
}

func checkPid(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func removePid(filepath string) error {
	zap.S().Debugf("Removeing the pid file %s", filepath)
	return os.Remove(filepath)
}

func readPid(filepath string) (int, error) {
	zap.S().Debugf("Trying to read the pid file %s", filepath)
	data, err := os.ReadFile(filepath)
	if err != nil {
		zap.S().Error("Failed reading the pid file")
		return -1, err
	}
	zap.S().Debugf("Success reading the pid file")

	zap.S().Debugf("Parsing data %s", string(data))
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		zap.S().Error("Failed to parse data")
		return -1, err
	}

	zap.S().Debugf("Success parsing data")
	return pid, nil
}
