package objstor

import (
	"context"
	"io"

	"github.com/killi1812/go-cache-server/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type mStorage struct {
	c *minio.Client
}

// DeleteFile implements ObjectStorage.
func (m *mStorage) DeleteFile(name string) error {
	panic("unimplemented")
}

// ReadFile implements ObjectStorage.
func (m *mStorage) ReadFile(name string) (io.ReadCloser, error) {
	panic("unimplemented")
}

// createDir implements ObjectStorage.
func (m *mStorage) createDir(name string) error {
	panic("unimplemented")
}

// createFile implements ObjectStorage.
func (m *mStorage) createFile(path string) error {
	panic("unimplemented")
}

// New creates a new minio.Client
func newMinioStorage() *mStorage {
	minioClient, err := minio.New(config.Config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Config.Minio.CredID, config.Config.Minio.CredSecret, config.Config.Minio.CredToken),
		Secure: config.Config.Minio.UseSSL,
	})
	if err != nil {
		zap.S().Panicf("failed to create MinIO client %w", err)
	}

	zap.S().Info("Pinging MinIO instance")
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		zap.S().DPanicf("Error pinging minio service, err: %w", err)
		return nil
	}
	zap.S().Debugf("MinIO contains %d buckets", len(buckets))

	return &mStorage{c: minioClient}
}
