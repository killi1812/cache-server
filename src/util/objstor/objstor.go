package objstor

import "os"

// common interface between OS file system and minio storage
type ObjectStorage interface {
	createDir(name string) error
	createFile(path string) error

	DeleteFile(name string) error
	// TODO: change to different interface of file
	ReadFile(name string) (os.File, error)
}
