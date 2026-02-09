package handler

import (
	"catalog-service/internal/repository"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CatalogService interface {
	CreateCatalogItem(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error)
	GetAllCatalogItems(ctx context.Context) ([]repository.Item, error)
	GetCatalogItemByID(ctx context.Context, id pgtype.UUID) (repository.Item, error)
	UpdateItemQuantity(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error)
	GetPresignedUploadURL(ctx context.Context, fileName string, fileType string) (string, string, error)
}

type CatalogHandler struct {
	service CatalogService
	logger  *slog.Logger
}

func NewHandler(service CatalogService, logger *slog.Logger) *CatalogHandler {
	return &CatalogHandler{
		service: service,
		logger:  logger,
	}
}

func CatalogRouter(r chi.Router, handler *CatalogHandler) {
	r.Get("/", handler.GetAllCatalogItems)
	r.Post("/", handler.CreateCatalogItem)

	
	r.Post("/upload-url", handler.GetPresignedUploadURL)
	r.Get("/{id}", handler.GetCatalogItemByID)
}
