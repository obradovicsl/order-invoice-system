package server

import (
	"catalog-service/internal/config"
	"catalog-service/internal/features/catalog"
	"catalog-service/internal/features/catalog/handler"
	"catalog-service/internal/middleware"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(config config.Config, catalogService *catalog.Service, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.CorsMiddleware(config.AllowedOrigins))
	router.Use(middleware.RecoveryMiddleware(logger))

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/catalog", func(r chi.Router) {
			cataloghandler := handler.NewHandler(catalogService, logger)
			handler.CatalogRouter(r, cataloghandler)
		})
	})

	return router
}
