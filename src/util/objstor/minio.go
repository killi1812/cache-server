package objstor

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/killi1812/go-cache-server/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

const (
	_DEFAULT_BUCKET = "cache-server"
	_LINK_MIME      = "application/x-nix-link"
)

type mStorage struct {
	c      *minio.Client
	bucket string
}

type progressTracker struct {
	io.Reader
	objectName string
	total      int64
	current    int64
	lastLog    int64
}

func (p *progressTracker) Read(b []byte) (int, error) {
	n, err := p.Reader.Read(b)
	p.current += int64(n)
	// Log every 10MB
	if p.current-p.lastLog >= 10*1024*1024 || p.current == p.total {
		zap.S().Infof("MinIO Uploading '%s': %d/%d bytes", p.objectName, p.current, p.total)
		p.lastLog = p.current
	}
	return n, err
}

// DeleteFile implements ObjectStorage.
func (m *mStorage) DeleteFile(cachename, name string) error {
	objectName := fmt.Sprintf("%s/%s", cachename, name)
	zap.S().Infof("MinIO: Trying to remove object '%s' from bucket '%s'", objectName, m.bucket)

	// If it's a link, we should probably delete the link but what about the data?
	// For now, simple delete of the named object.
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
	ctx := context.Background()

	// Check if it's a pointer/link
	info, err := m.c.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err == nil && info.ContentType == _LINK_MIME {
		zap.S().Infof("MinIO: Following link '%s'", objectName)
		obj, err := m.c.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
		if err == nil {
			data, _ := io.ReadAll(obj)
			obj.Close()
			objectName = fmt.Sprintf("%s/%s", cachename, string(data))
		}
	}

	zap.S().Infof("MinIO: Trying to read object '%s' from bucket '%s'", objectName, m.bucket)
	return m.c.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
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
func (m *mStorage) WriteFile(cachename, name string, data io.Reader, size int64) error {
	objectName := fmt.Sprintf("%s/%s", cachename, name)
	zap.S().Infof("MinIO: Trying to write object '%s' to bucket '%s'", objectName, m.bucket)

	pt := &progressTracker{
		Reader:     data,
		objectName: objectName,
		total:      size,
	}

	_, err := m.c.PutObject(context.Background(), m.bucket, objectName, pt, size, minio.PutObjectOptions{
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

// RenameFile implements ObjectStorage.
func (m *mStorage) RenameFile(cachename, oldName, newName string) error {
	ctx := context.Background()
	oldObjectName := fmt.Sprintf("%s/%s", cachename, oldName)
	newObjectName := fmt.Sprintf("%s/%s", cachename, newName)

	zap.S().Infof("MinIO: Creating link object '%s' -> '%s' in bucket '%s'", newObjectName, oldObjectName, m.bucket)

	_, err := m.c.PutObject(ctx, m.bucket, newObjectName, strings.NewReader(oldName), int64(len(oldName)), minio.PutObjectOptions{
		ContentType: _LINK_MIME,
	})
	if err != nil {
		zap.S().Errorf("MinIO: Failed to create link object: %v", err)
		return err
	}

	return nil
}

// Stat implements ObjectStorage.
func (m *mStorage) Stat(cachename, name string) (int64, error) {
	objectName := fmt.Sprintf("%s/%s", cachename, name)
	ctx := context.Background()
	info, err := m.c.StatObject(ctx, m.bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}

	if info.ContentType == _LINK_MIME {
		obj, err := m.c.GetObject(ctx, m.bucket, objectName, minio.GetObjectOptions{})
		if err == nil {
			data, _ := io.ReadAll(obj)
			obj.Close()
			targetName := fmt.Sprintf("%s/%s", cachename, string(data))
			targetInfo, err := m.c.StatObject(ctx, m.bucket, targetName, minio.StatObjectOptions{})
			if err == nil {
				return targetInfo.Size, nil
			}
		}
	}

	return info.Size, nil
}

func newMinioStorage() *mStorage {
	minioClient, err := minio.New(config.Config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Config.Minio.CredID, config.Config.Minio.CredSecret, config.Config.Minio.CredToken),
		Secure: config.Config.Minio.UseSSL,
	})
	if err != nil {
		zap.S().Panicf("failed to create MinIO client %w", err)
		return nil
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
