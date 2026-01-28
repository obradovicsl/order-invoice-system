package catalog

import (
	"catalog-service/internal/repository"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	CreateCatalogItem(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error)
	GetAllCatalogItems(ctx context.Context) ([]repository.Item, error)
	GetCatalogItemByID(ctx context.Context, id pgtype.UUID) (repository.Item, error)
	UpdateItemQuantity(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error)
}

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) CreateCatalogItem(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error) {
	return s.repo.CreateCatalogItem(ctx, arg)
}

func (s *Service) GetAllCatalogItems(ctx context.Context) ([]repository.Item, error) {
	return s.repo.GetAllCatalogItems(ctx)
}

func (s *Service) GetCatalogItemByID(ctx context.Context, id pgtype.UUID) (repository.Item, error) {
	return s.repo.GetCatalogItemByID(ctx, id)
}

func (s *Service) UpdateItemQuantity(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error) {
	return s.repo.UpdateItemQuantity(ctx, arg)
}
