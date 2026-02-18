package objstor

import "os"

// common interface between OS file system and minio storage
type ObjectStorage interface {
	CreateDir(name string) error
	CreateFile(name string) error

	DeleteFile(name string) error
	// TODO: change to different interface of file
	ReadFile(name string) (os.File, error)
}

func New() ObjectStorage {
	return newFileStorage()
}
