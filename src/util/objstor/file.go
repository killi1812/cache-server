package objstor

import "os"

// 3. Create the physical directory for store paths
// Path example: ./data/caches/<cache_name>

// cachePath := filepath.Join(m.RootDir, name)
// if err := os.MkdirAll(cachePath, 0o755); err != nil {
// 	return fmt.Errorf("failed to create cache directory at %s: %w", cachePath, err)
// }
// zap.S().Debugf("Created storage directory: %s", cachePath)
//
//

type fileStorage struct {
	workingDir string
}

// DeleteFile implements ObjectStorage.
func (f fileStorage) DeleteFile(name string) error {
	panic("unimplemented")
}

// ReadFile implements ObjectStorage.
func (f fileStorage) ReadFile(name string) (os.File, error) {
	panic("unimplemented")
}

// CreateDir implements ObjectStorage.
func (f fileStorage) CreateDir(name string) error {
	panic("unimplemented")
}

// CreateFile implements ObjectStorage.
func (f fileStorage) CreateFile(path string) error {
	panic("unimplemented")
}

func newFileStorage() ObjectStorage {
	return fileStorage{}
}
