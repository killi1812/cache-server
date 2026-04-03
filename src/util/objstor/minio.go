package objstor

import (
	"context"
	"fmt"
	"io"

	"github.com/killi1812/go-cache-server/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

const _DEFAULT_BUCKET = "cache-server"

type mStorage struct {
	c      *minio.Client
	bucket string
}

// DeleteFile implements ObjectStorage.
func (m *mStorage) DeleteFile(cachename, name string) error {
	objectName := fmt.Sprintf("%s/%s", cachename, name)
	zap.S().Infof("MinIO: Trying to remove object '%s' from bucket '%s'", objectName, m.bucket)
	err := m.c.RemoveObject(context.Background(), m.bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		zap.S().Errorf("MinIO: Failed to remove object '%s', err: %v", objectName, err)
		return err
	}
	return nil
}

// ReadFile implements ObjectStorage.
func (m *mStorage) ReadFile(cachename, name string) (io.ReadCloser, error) {
	objectName := fmt.Sprintf("%s/%s", cachename, name)
	zap.S().Infof("MinIO: Trying to read object '%s' from bucket '%s'", objectName, m.bucket)
	return m.c.GetObject(context.Background(), m.bucket, objectName, minio.GetObjectOptions{})
}

// CreateDir implements ObjectStorage.
func (m *mStorage) CreateDir(name string) (string, error) {
	// In MinIO, we ensure the bucket exists. 'name' could be used as a prefix if needed,
	// but here we just ensure the main bucket exists.
	ctx := context.Background()
	exists, err := m.c.BucketExists(ctx, m.bucket)
	if err != nil {
		return "", err
	}
	if !exists {
		zap.S().Infof("MinIO: Creating bucket '%s'", m.bucket)
		err = m.c.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return "", err
		}
	}
	return m.bucket, nil
}

// WriteFile implements ObjectStorage.
func (m *mStorage) WriteFile(cachename, name string, data io.Reader) error {
	objectName := fmt.Sprintf("%s/%s", cachename, name)
	zap.S().Infof("MinIO: Trying to write object '%s' to bucket '%s'", objectName, m.bucket)
	// Using -1 for size tells minio-go to use internal buffering for unknown size
	_, err := m.c.PutObject(context.Background(), m.bucket, objectName, data, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		zap.S().Errorf("MinIO: Failed to write object '%s', err: %v", objectName, err)
		return err
	}
	return nil
}

// CreatFile implements ObjectStorage.
func (m *mStorage) CreatFile(cachename, filename string) error {
	objectName := fmt.Sprintf("%s/%s", cachename, filename)
	zap.S().Infof("MinIO: Creating placeholder object '%s' in bucket '%s'", objectName, m.bucket)

	// Create an empty object
	_, err := m.c.PutObject(context.Background(), m.bucket, objectName, nil, 0, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		zap.S().Errorf("MinIO: Failed to create object '%s', err: %v", objectName, err)
		return err
	}
	return nil
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

	return &mStorage{c: minioClient, bucket: _DEFAULT_BUCKET}
}
