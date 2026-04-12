package objstor

import (
	"testing"

	"github.com/killi1812/go-cache-server/config"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config.Config = config.NewConfig()

	t.Run("File Storage", func(t *testing.T) {
		config.Config.CacheServer.StorageType = "file"
		storage := New()
		assert.NotNil(t, storage)
		_, ok := storage.(fileStorage)
		assert.True(t, ok)
	})

	// TODO: test minio storages mby should return errors rather then nil no not being able to ping test storage
	//
	// t.Run("Minio Storage", func(t *testing.T) {
	// 	config.Config.CacheServer.StorageType = "minio"
	// 	storage := New()
	// 	if storage == nil {
	// 		t.Skip("Minio service not available")
	// 	}
	// 	assert.NotNil(t, storage)
	// })
	//
	// t.Run("Factory produces same objects minio", func(t *testing.T) {
	// 	config.Config.CacheServer.StorageType = "minio"
	// 	minio := newMinioStorage()
	// 	storage := New()
	//
	// 	if storage == nil {
	// 		t.Skip("Minio service not available")
	// 	}
	// 	assert.NotNil(t, storage)
	// 	assert.NotNil(t, minio)
	// 	assert.Equal(t, minio, storage)
	// })
	//
	// t.Run("Factory produces same objects file", func(t *testing.T) {
	// 	config.Config.CacheServer.StorageType = "file"
	// 	fstorage := newFileStorage(config.Config.CacheServer.CacheDir)
	// 	storage := New()
	//
	// 	assert.NotNil(t, storage)
	// 	assert.NotNil(t, fstorage)
	// 	assert.Equal(t, fstorage, storage)
	// })
}
