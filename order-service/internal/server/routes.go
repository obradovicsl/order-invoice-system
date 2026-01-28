package server

import (
	"log/slog"
	"net/http"
	"order-service/internal/config"
	"order-service/internal/features/order"
	"order-service/internal/features/order/handler"
	"order-service/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter(config config.Config, orderService *order.Service, logger *slog.Logger) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.CorsMiddleware(config.AllowedOrigins))
	router.Use(middleware.RecoveryMiddleware(logger))

	router.Route("/api/v1", func(r chi.Router) {
		r.Route("/orders", func(r chi.Router) {
			orderHandler := handler.NewHandler(orderService, logger)
			handler.OrderRouter(r, orderHandler)
		})
	})

	return router
}
