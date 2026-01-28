package handler

import (
	"catalog-service/internal/repository"
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CatalogService definiše interfejs koji implementira servis
type CatalogService interface {
	CreateCatalogItem(ctx context.Context, arg repository.CreateCatalogItemParams) (repository.Item, error)
	GetAllCatalogItems(ctx context.Context) ([]repository.Item, error)
	GetCatalogItemByID(ctx context.Context, id pgtype.UUID) (repository.Item, error)
	UpdateItemQuantity(ctx context.Context, arg repository.UpdateItemQuantityParams) (repository.Item, error)
}

// Handler sadrži servis i logiku za HTTP zahtjeve
type CatalogHandler struct {
	service CatalogService
	logger  *slog.Logger
}

// NewHandler kreira novi handler
func NewHandler(service CatalogService, logger *slog.Logger) *CatalogHandler {
	return &CatalogHandler{
		service: service,
		logger:  logger,
	}
}

func CatalogRouter(r chi.Router, handler *CatalogHandler) {
	r.Get("/", handler.GetAllCatalogItems)
	r.Get("/{id}", handler.GetCatalogItemByID)
	r.Post("/", handler.CreateCatalogItem)
}
