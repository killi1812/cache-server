package objstor

import (
	"errors"
	"io"
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
	// TODO: change to not tmp
	return newFileStorage("tmp/cache")
}

// errors

var ErrFailedToCreateDir = errors.New("error failed to create dir")
