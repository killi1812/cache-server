package objstor

import (
	"context"

	"github.com/killi1812/go-cache-server/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// New creates a new minio.Client
func NewMinio() *minio.Client {
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

	return minioClient
}
