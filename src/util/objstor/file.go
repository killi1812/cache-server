package objstor

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

type fileStorage struct {
	rootDir string
}

const filePerms = 0o755

// DeleteFile implements ObjectStorage.
func (f fileStorage) DeleteFile(name string) error {
	panic("unimplemented")
}

// ReadFile implements ObjectStorage.
func (f fileStorage) ReadFile(name string) (os.File, error) {
	panic("unimplemented")
}

// CreateDir implements ObjectStorage.
func (f fileStorage) CreateDir(name string) (string, error) {
	cachePath := filepath.Join(f.rootDir, name)
	if err := os.MkdirAll(cachePath, filePerms); err != nil {
		zap.S().Errorf("Failed to create cache directory at %s: %w", cachePath, err)
		return "", ErrFailedToCreateDir
	}
	zap.S().Debugf("Created storage directory: %s", cachePath)

	return cachePath, nil
}

// CreateFile implements ObjectStorage.
func (f fileStorage) CreateFile(path string) error {
	panic("unimplemented")
}

func newFileStorage(rootDir string) ObjectStorage {
	return fileStorage{rootDir}
}
