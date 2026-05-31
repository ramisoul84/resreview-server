package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ramisoul84/resreview-server/internal/domain"
	"github.com/ramisoul84/resreview-server/internal/repository"
	"github.com/ramisoul84/resreview-server/pkg/logger"
)

type ProductService interface {
	CreateProduct(ctx context.Context, userID, name string) (*domain.ProductResponse, error)
	GetProduct(ctx context.Context, productID, userID string) (*domain.ProductResponse, error)
	ListProducts(ctx context.Context, userID string) ([]domain.ProductResponse, error)
	ListAllProducts(ctx context.Context) ([]domain.ProductResponse, error)
	ListAllProductsWithVersions(ctx context.Context) ([]domain.ProductWithVersionsResponse, error)
	UpdateProduct(ctx context.Context, productID, userID, name string) error
	DeleteProduct(ctx context.Context, productID, userID string) error
}

type productService struct {
	repo        repository.ProductRepository
	versionRepo repository.VersionRepository
	log         logger.Logger
}

func NewProductService(repo repository.ProductRepository, versionRepo repository.VersionRepository) ProductService {
	return &productService{repo: repo, versionRepo: versionRepo, log: logger.Get()}
}

func (s *productService) CreateProduct(ctx context.Context, userID, name string) (*domain.ProductResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "CreateProduct",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("creating product")

	product := &domain.Product{
		ID:        uuid.New().String(),
		Name:      name,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		log.Error().Err(err).Msg("failed to create product")
		return nil, err
	}

	return &domain.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		UserID:    product.UserID,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *productService) GetProduct(ctx context.Context, productID, userID string) (*domain.ProductResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "GetProduct",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("getting product")

	product, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	return &domain.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		UserID:    product.UserID,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *productService) ListProducts(ctx context.Context, userID string) ([]domain.ProductResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "ListProducts",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("user_id", userID).Msg("listing products")

	products, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list products")
		return nil, err
	}

	return products, nil
}

func (s *productService) ListAllProducts(ctx context.Context) ([]domain.ProductResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "ListAllProducts",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("listing all products")

	products, err := s.repo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list all products")
		return nil, err
	}

	return products, nil
}

func (s *productService) ListAllProductsWithVersions(ctx context.Context) ([]domain.ProductWithVersionsResponse, error) {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "ListAllProductsWithVersions",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Msg("listing all products with versions")

	products, err := s.repo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list all products")
		return nil, err
	}

	versions, err := s.versionRepo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list all versions")
		return nil, err
	}

	versionMap := make(map[string][]domain.VersionResponse, len(products))
	for _, v := range versions {
		versionMap[v.ProductID] = append(versionMap[v.ProductID], v)
	}

	result := make([]domain.ProductWithVersionsResponse, 0, len(products))
	for _, p := range products {
		pVersions := versionMap[p.ID]
		if pVersions == nil {
			pVersions = []domain.VersionResponse{}
		}
		result = append(result, domain.ProductWithVersionsResponse{
			ID:        p.ID,
			Name:      p.Name,
			UserID:    p.UserID,
			CreatedAt: p.CreatedAt,
			Versions:  pVersions,
		})
	}

	log.Debug().Int("count", len(result)).Msg("all products with versions fetched successfully")
	return result, nil
}

func (s *productService) UpdateProduct(ctx context.Context, productID, userID, name string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "UpdateProduct",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("updating product")

	return s.repo.Update(ctx, productID, userID, name)
}

func (s *productService) DeleteProduct(ctx context.Context, productID, userID string) error {
	log := s.log.WithFields(map[string]any{
		"layer":      "product_service",
		"method":     "DeleteProduct",
		"request_id": ctx.Value("request_id"),
	})

	log.Debug().Str("product_id", productID).Msg("deleting product")

	return s.repo.Delete(ctx, productID, userID)
}
