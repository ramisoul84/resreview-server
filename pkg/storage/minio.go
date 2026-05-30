package storage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ramisoul84/resreview-server/internal/config"
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// PhotoStorage handles upload and deletion of photo objects in MinIO.
type PhotoStorage interface {
	Upload(ctx context.Context, versionID string, data []byte, mimeType string) (url, key string, err error)
	Delete(ctx context.Context, key string) error
}

type minioStorage struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewMinIOStorage(cfg config.MinIOConfig) (PhotoStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: create client: %w", err)
	}

	stor := &minioStorage{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}

	if cfg.InitBucketOnStart {
		if err := stor.ensureBucket(context.Background()); err != nil {
			return nil, err
		}
	}

	return stor, nil
}

func (s *minioStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("minio: check bucket %q: %w", s.bucket, err)
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("minio: create bucket %q: %w", s.bucket, err)
		}
	}

	// Public-read policy — GET requests on any object need no credentials.
	// This lets Nginx proxy /photos/ → minio:9000/<bucket>/ without auth headers.
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect":    "Allow",
			"Principal": {"AWS": ["*"]},
			"Action":    ["s3:GetObject"],
			"Resource":  ["arn:aws:s3:::%s/*"]
		}]
	}`, s.bucket)

	if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
		return fmt.Errorf("minio: set bucket policy: %w", err)
	}

	return nil
}

func (s *minioStorage) Upload(ctx context.Context, versionID string, data []byte, mimeType string) (string, string, error) {
	ext, ok := allowedMimeTypes[mimeType]
	if !ok {
		return "", "", fmt.Errorf("minio: unsupported mime type: %s", mimeType)
	}

	// Key layout: photos/{userID}/{uuid}.ext
	// Partitioned by versionID so per-user listing is efficient.
	key := fmt.Sprintf("photos/%s/%s%s", versionID, uuid.NewString(), ext)

	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		key,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: mimeType},
	)
	if err != nil {
		return "", "", fmt.Errorf("minio: upload: %w", err)
	}

	// URL = publicURL/key  e.g. https://ramisuliman.ru/photos/dating-photos/photos/{userID}/{uuid}.jpg
	url := fmt.Sprintf("%s/%s", s.publicURL, key)
	return url, key, nil
}

func (s *minioStorage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio: delete %q: %w", key, err)
	}
	return nil
}
