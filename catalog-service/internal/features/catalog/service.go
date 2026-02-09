package catalog

import (
	"catalog-service/internal/repository"
	"catalog-service/internal/service"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	CreateCatalogItem(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error)
	GetAllCatalogItems(ctx context.Context) ([]repository.Item, error)
	GetCatalogItemByID(ctx context.Context, id pgtype.UUID) (repository.Item, error)
	UpdateItemQuantity(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error)
}

type Service struct {
	repo         Repository
	logger       *slog.Logger
	blobService  *service.BlobService
	blobContName string
}

func NewService(repo Repository, logger *slog.Logger, blobService *service.BlobService, blobContName string) *Service {
	return &Service{
		repo:         repo,
		logger:       logger,
		blobService:  blobService,
		blobContName: blobContName,
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

// GetPresignedUploadURL generates a presigned URL for direct blob upload
func (s *Service) GetPresignedUploadURL(ctx context.Context, fileName string, fileType string) (string, error) {
	s.logger.Info("generating presigned upload URL",
		"file_name", fileName,
		"file_type", fileType,
	)

	// Generate signed URL with 1-hour expiry for upload
	signedURL, err := s.blobService.GenerateUploadSignedURL(s.blobContName, fileName, 1*time.Hour)
	if err != nil {
		s.logger.Error("failed to generate presigned upload URL",
			"file_name", fileName,
			"error", err,
		)
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	s.logger.Info("presigned upload URL generated successfully", "file_name", fileName)
	return signedURL, nil
}
