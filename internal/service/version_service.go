package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
	"github.com/ramisoul84/resreview-server/pkg/storage"
)

type VersionService interface {
	CreateVersion(ctx context.Context, productID, userID, label, name string) (*domain.VersionResponse, error)
	GetVersion(ctx context.Context, versionID, productID string) (*domain.VersionResponse, error)
	ListVersions(ctx context.Context, productID string) ([]domain.VersionResponse, error)
	UpdateVersion(ctx context.Context, versionID, productID, userID, label, name string) error
	DeleteVersion(ctx context.Context, versionID, productID, userID string) error
	UploadVersionImage(ctx context.Context, versionID, productID, userID string, data []byte, mimeType string) (*domain.VersionResponse, error)
}

type versionService struct {
	versionRepo repository.VersionRepository
	productRepo repository.ProductRepository
	photoRepo   repository.PhotoRepository
	storage     storage.PhotoStorage
	maxSize     int64
	log         logger.Logger
}

func NewVersionService(
	versionRepo repository.VersionRepository,
	productRepo repository.ProductRepository,
	photoRepo repository.PhotoRepository,
	storage storage.PhotoStorage,
	maxSize int64) VersionService {
	return &versionService{
		versionRepo: versionRepo,
		photoRepo:   photoRepo,
		productRepo: productRepo,
		storage:     storage,
		maxSize:     maxSize,
		log:         logger.Get(),
	}
}

func (s *versionService) verifyProductOwnership(ctx context.Context, productID, userID string) error {
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.UserID != userID {
		return domain.ErrProductNotFound
	}
	return nil
}

func (s *versionService) CreateVersion(ctx context.Context, productID, userID, label, name string) (*domain.VersionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "version_service",
		"method":     "CreateVersion",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("creating version")

	if err := s.verifyProductOwnership(ctx, productID, userID); err != nil {
		return nil, err
	}

	version := &domain.Version{
		ID:        uuid.New().String(),
		Label:     label,
		Name:      name,
		ProductID: productID,
		CreatedAt: time.Now(),
	}

	if err := s.versionRepo.Create(ctx, version); err != nil {
		log.Error().Err(err).Msg("failed to create version")
		return nil, err
	}

	return &domain.VersionResponse{
		ID:        version.ID,
		Label:     version.Label,
		Name:      version.Name,
		ProductID: version.ProductID,
		URL:       version.URL,
		CreatedAt: version.CreatedAt,
	}, nil
}

func (s *versionService) GetVersion(ctx context.Context, versionID, productID string) (*domain.VersionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "version_service",
		"method":     "GetVersion",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("getting version")

	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, err
	}

	if version.ProductID != productID {
		return nil, domain.ErrVersionNotFound
	}

	return &domain.VersionResponse{
		ID:        version.ID,
		Label:     version.Label,
		Name:      version.Name,
		ProductID: version.ProductID,
		URL:       version.URL,
		CreatedAt: version.CreatedAt,
	}, nil
}

func (s *versionService) ListVersions(ctx context.Context, productID string) ([]domain.VersionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "version_service",
		"method":     "ListVersions",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("listing versions")

	versions, err := s.versionRepo.ListByProductID(ctx, productID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list versions")
		return nil, err
	}

	return versions, nil
}

func (s *versionService) UpdateVersion(ctx context.Context, versionID, productID, userID, label, name string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "version_service",
		"method":     "UpdateVersion",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("updating version")

	if err := s.verifyProductOwnership(ctx, productID, userID); err != nil {
		return err
	}

	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}

	if version.ProductID != productID {
		return domain.ErrVersionNotFound
	}

	return s.versionRepo.Update(ctx, versionID, label, name)
}

func (s *versionService) DeleteVersion(ctx context.Context, versionID, productID, userID string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "version_service",
		"method":     "DeleteVersion",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("deleting version")

	if err := s.verifyProductOwnership(ctx, productID, userID); err != nil {
		return err
	}

	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return err
	}

	if version.ProductID != productID {
		return domain.ErrVersionNotFound
	}

	return s.versionRepo.Delete(ctx, versionID)
}

func (s *versionService) UploadVersionImage(ctx context.Context, versionID, productID, userID string, data []byte, mimeType string) (*domain.VersionResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "version_service",
		"method":     "UploadVersionImage",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("version_id", versionID).Msg("uploading version image")

	if err := s.verifyProductOwnership(ctx, productID, userID); err != nil {
		return nil, err
	}

	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, err
	}
	if version.ProductID != productID {
		return nil, domain.ErrVersionNotFound
	}

	if _, ok := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/webp": true,
	}[mimeType]; !ok {
		return nil, domain.ErrInvalidMimeType
	}
	if int64(len(data)) > s.maxSize {
		return nil, domain.ErrFileTooLarge
	}

	url, key, err := s.storage.Upload(ctx, versionID, data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("version_svc.UploadVersionImage: upload: %w", err)
	}

	if err := s.versionRepo.UpdateURL(ctx, versionID, url); err != nil {
		_ = s.storage.Delete(ctx, key)
		return nil, fmt.Errorf("version_svc.UploadVersionImage: update url: %w", err)
	}

	version.URL = url

	log.Info().Str("version_id", versionID).Str("url", url).Msg("version image uploaded")

	return &domain.VersionResponse{
		ID:        version.ID,
		Label:     version.Label,
		Name:      version.Name,
		ProductID: version.ProductID,
		URL:       version.URL,
		CreatedAt: version.CreatedAt,
	}, nil
}
