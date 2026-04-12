package proc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/killi1812/go-cache-server/util/pid"
	"github.com/stretchr/testify/assert"
)

func TestProc(t *testing.T) {
	// The pid package hardcodes /tmp as the base directory.
	// We use a unique filename to avoid conflicts.
	pidFile := "test-proc-" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".pid"
	
	// Ensure we cleanup at the end
	defer os.Remove(filepath.Join("/tmp", pidFile))

	t.Run("IsRunning - Not Running", func(t *testing.T) {
		assert.False(t, IsRunning(pidFile))
	})

	t.Run("StopProc - No Pid File", func(t *testing.T) {
		err := StopProc(pidFile)
		assert.Error(t, err)
		assert.Equal(t, ErrFailedToStop, err)
	})

	t.Run("StopProc - Success", func(t *testing.T) {
		// Start a dummy process
		cmd := exec.Command("sleep", "10")
		err := cmd.Start()
		assert.NoError(t, err)

		// Write its PID to file
		err = pid.WritePid(pidFile, cmd.Process.Pid)
		assert.NoError(t, err)

		assert.True(t, IsRunning(pidFile))

		// Stop it
		err = StopProc(pidFile)
		assert.NoError(t, err)

		// Wait a bit for signal to be processed
		time.Sleep(100 * time.Millisecond)

		// Check if it's still running (it shouldn't be, or at least the pid file should be gone)
		// readPid in StopProc removes the pid file
		_, err = os.Stat(pidFile)
		assert.True(t, os.IsNotExist(err))
		
		// Cleanup
		cmd.Process.Kill()
	})
}
