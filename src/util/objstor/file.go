package objstor

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type fileStorage struct {
	rootDir string
}

const (
	dirPerms  = 0o755
	filePerms = 0o644
)

// WriteFile implements [ObjectStorage].
func (f fileStorage) WriteFile(name string, data io.Reader) error {
	zap.S().Infof("Trying to write file '%s'", name)
	cachePath := filepath.Join(f.rootDir, name)

	// TODO: missing cache name
	file, err := os.OpenFile(cachePath, os.O_WRONLY, filePerms)
	if err != nil {
		zap.S().Errorf("Failed to open file '%s', err: %v", cachePath, err)
		return err
	}
	defer file.Close()

	size, err := io.Copy(file, data)
	if err != nil {
		zap.S().Errorf("Failed to write data to file, err: %v", err)
		return err
	}

	zap.S().Infof("Successfuly writen %d bytes of data", size)
	return nil
}

// DeleteFile implements ObjectStorage.
func (f fileStorage) DeleteFile(name string) error {
	zap.S().Infof("Trying to remove file '%s'", name)
	cachePath := filepath.Join(f.rootDir, name)

	info, err := os.Stat(cachePath)
	if err != nil {
		pErr := err.(*os.PathError)
		zap.S().Errorf("Failed to access path 's', err: %+v", pErr)
	}
	zap.S().Infof("File info %+v", info)

	err = os.Remove(cachePath)
	if err != nil {
		pErr := err.(*os.PathError)
		zap.S().Errorf("Failed to remove path '%s', err: %+v", cachePath, pErr)
		return err
	}

	zap.S().Infof("File '%s' removed successfuly!", name)
	return nil
}

// ReadFile implements ObjectStorage.
func (f fileStorage) ReadFile(name string) (io.ReadCloser, error) {
	cachePath := filepath.Join(f.rootDir, name)
	return os.Open(cachePath)
}

// CreateDir implements ObjectStorage.
func (f fileStorage) CreateDir(name string) (string, error) {
	cachePath := filepath.Join(f.rootDir, name)
	if err := os.MkdirAll(cachePath, dirPerms); err != nil {
		zap.S().Errorf("Failed to create cache directory at %s: %w", cachePath, err)
		return "", ErrFailedToCreateDir
	}
	zap.S().Debugf("Created storage directory: %s", cachePath)

	return cachePath, nil
}

func (f fileStorage) CreatFile(cachename, filename string) error {
	zap.S().Infof("Trying to write file '%s'", filename)
	cachePath := filepath.Join(f.rootDir, cachename, filename)

	file, err := os.OpenFile(cachePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filePerms)
	if err != nil {
		zap.S().Errorf("Failed to create file '%s', err: %v", cachePath, err)
		return err
	}
	defer file.Close()

	zap.S().Infof("File Created successfulyy")
	return nil
}

func newFileStorage(rootDir string) ObjectStorage {
	return fileStorage{rootDir}
}
