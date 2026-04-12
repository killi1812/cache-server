package objstor

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storage := newFileStorage(tmpDir)

	t.Run("CreateDir", func(t *testing.T) {
		path, err := storage.CreateDir("test-cache")
		assert.NoError(t, err)
		assert.Contains(t, path, "test-cache")
		
		info, err := os.Stat(path)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("WriteFile and ReadFile", func(t *testing.T) {
		cacheName := "test-cache"
		fileName := "test-file"
		content := []byte("hello world")
		
		// Ensure dir exists
		_, err := storage.CreateDir(cacheName)
		assert.NoError(t, err)

		err = storage.WriteFile(cacheName, fileName, bytes.NewBuffer(content))
		assert.NoError(t, err)

		reader, err := storage.ReadFile(cacheName, fileName)
		assert.NoError(t, err)
		defer reader.Close()

		readContent, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, content, readContent)
	})

	t.Run("CreatFile", func(t *testing.T) {
		cacheName := "test-cache"
		fileName := "empty-file"
		
		err := storage.CreatFile(cacheName, fileName)
		assert.NoError(t, err)

		info, err := os.Stat(filepath.Join(tmpDir, cacheName, fileName))
		assert.NoError(t, err)
		assert.Equal(t, int64(0), info.Size())
	})

	t.Run("DeleteFile", func(t *testing.T) {
		cacheName := "test-cache"
		fileName := "delete-me"
		
		storage.CreatFile(cacheName, fileName)
		
		err := storage.DeleteFile(cacheName, fileName)
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, cacheName, fileName))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("ReadFile - Not Found", func(t *testing.T) {
		_, err := storage.ReadFile("nonexistent", "file")
		assert.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("DeleteFile - Not Found", func(t *testing.T) {
		err := storage.DeleteFile("nonexistent", "file")
		assert.Error(t, err)
	})
}
