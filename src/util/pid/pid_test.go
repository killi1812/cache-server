package pid

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

const _EXAMPLE_PID = 1028931

func TestPidFileExists(t *testing.T) {
	filename := "testdata/example.pid"

	// checkPid
	assert.True(t, checkPid(filename), "Pid file should exist")
	// readPid
	data, err := readPid(filename)
	assert.Nil(t, err, "ReadPid err should be nil")
	assert.NotEqual(t, -1, data, "Data should not be -1")
	assert.Equal(t, _EXAMPLE_PID, data, "Data should not be equal to %s")
}

func TestReadWritePid(t *testing.T) {
	// writePid
	filename := "testdata/tmp.pid"
	assert.Nil(t, writePid(_EXAMPLE_PID, filename), "WritePid err should be nil")

	err := writePid(_EXAMPLE_PID, filename)
	assert.Error(t, err, "WritePid err should be not nil")
	assert.ErrorIs(t, err, os.ErrExist, "WritePid err should be bad permissions")

	t.Cleanup(func() { removePid(filename) })
	// check if exits
	assert.True(t, checkPid(filename), "Pid file should exist")

	// check if pid correct
	data, err := readPid(filename)
	assert.Nil(t, err, "ReadPid err should be nil")
	assert.NotEqual(t, -1, data, "Data should not be -1")
	assert.Equal(t, _EXAMPLE_PID, data, "Data should not be equal to %s")

	// RemovePid
	assert.Nil(t, removePid(filename), "RemovePid err should be nil")
	// check if exits
	assert.False(t, checkPid(filename), "Pid file should not exist anymore")
}

func TestNoPidFile(t *testing.T) {
	filename := "testdata/no-name.pid"

	// checkPid
	assert.False(t, checkPid(filename), "Pid file should not exist")

	// readPid
	data, err := readPid(filename)
	assert.ErrorIs(t, err, os.ErrNotExist, "ReadPid err should be file ErrNotExist")
	assert.Equal(t, -1, data, "Data should be -1")

	assert.ErrorIs(t, removePid(filename), os.ErrNotExist, "RemovePid err should be ErrNotexist")
}

func TestBadData(t *testing.T) {
	filename := "testdata/bad_data.pid"

	// checkPid
	assert.True(t, checkPid(filename), "Pid file should exist")

	// readPid
	data, err := readPid(filename)
	assert.ErrorIs(t, err, strconv.ErrSyntax, "ReadPid err should be file ErrSyntax")
	assert.Equal(t, -1, data, "Data should be -1")
}

func TestBadPerms(t *testing.T) {
	filename := "testdata/bad_perms.pid"

	err := os.Chmod(filename, 0o000)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(filename, 0o644)

	// checkPid
	assert.True(t, checkPid(filename), "Pid file should exist")

	// readPid
	data, err := readPid(filename)
	assert.ErrorIs(t, err, os.ErrPermission, "ReadPid err should be file ErrSyntax")
	assert.Equal(t, -1, data, "Data should be -1")

	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "private")

	os.Mkdir(restrictedDir, 0o755)
	filename = filepath.Join(restrictedDir, filename)
	os.WriteFile(filename, []byte(strconv.Itoa(_EXAMPLE_PID)), 0o644)

	// Make the FOLDER unreadable/unwritable
	os.Chmod(restrictedDir, 0o000)
	defer os.Chmod(restrictedDir, 0o755)

	err = writePid(_EXAMPLE_PID, filename)
	assert.ErrorIs(t, err, os.ErrPermission, "WritePid err should be nil")
}
