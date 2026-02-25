package objstor

import (
	"errors"
	"io"
	"os"
)

// common interface between OS file system and minio storage
type ObjectStorage interface {
	CreateDir(name string) (string, error)
	WriteFile(name string, file io.Reader) error

	DeleteFile(name string) error
	// TODO: change to different interface of file
	ReadFile(name string) (os.File, error)
}

func New() ObjectStorage {
	// TODO: change to not tmp
	return newFileStorage("tmp/cache")
}

// errors

var ErrFailedToCreateDir = errors.New("error failed to create dir")
