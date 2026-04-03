package objstor

import (
	"errors"
	"io"

	"github.com/killi1812/go-cache-server/config"
)

// common interface between OS file system and minio storage
type ObjectStorage interface {
	CreateDir(name string) (string, error)
	WriteFile(name string, file io.Reader) error
	CreatFile(cachename, filename string) error

	DeleteFile(name string) error
	ReadFile(name string) (io.ReadCloser, error)
}

func New() ObjectStorage {
	if config.Config.CacheServer.StorageType == "minio" {
		return newMinioStorage()
	}
	return newFileStorage(config.Config.CacheServer.CacheDir)
}

// errors

var ErrFailedToCreateDir = errors.New("error failed to create dir")
