package handler

import (
	"context"
	"log/slog"
	"math"
	"math/big"
	"order-service/internal/features/order"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderService interface {
	GetAllOrders(ctx context.Context) ([]order.OrderResponse, error)
	GetOrderByID(ctx context.Context, id string) (*order.OrderResponse, error)
	CreateOrder(ctx context.Context, req order.CreateOrderRequest) (*order.OrderResponse, error)
}

type OrderHandler struct {
	service OrderService
	logger  *slog.Logger
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid {
		return 0
	}
	f := new(big.Float).SetInt(n.Int)
	f.Mul(f, big.NewFloat(math.Pow10(int(n.Exp))))
	result, _ := f.Float64()
	return result
}

func NewHandler(service OrderService, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

func OrderRouter(r chi.Router, handler *OrderHandler) {
	r.Get("/", handler.GetAllOrders)
	r.Get("/{id}", handler.GetOrderByID)
	r.Post("/", handler.CreateOrder)
}