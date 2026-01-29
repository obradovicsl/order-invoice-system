package service

import (
	"context"
	"fmt"
	"invoice-worker/internal/repository"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
)

type OrderRepository interface {
	GetOrderByID(ctx context.Context, id pgtype.UUID) ([]repository.GetOrderByIDRow, error)
	GetItemByID(ctx context.Context, id pgtype.UUID) (repository.GetItemByIDRow, error)
	UpdateOrderStatus(ctx context.Context, arg repository.UpdateOrderStatusParams) (repository.Order, error)
}

type OrderService struct {
	repo   OrderRepository
	logger *slog.Logger
}

type OrderItem struct {
	ID           pgtype.UUID    `json:"id"`
	ItemID       pgtype.UUID    `json:"item_id"`
	ItemCode     string         `json:"code"`
	ItemName     string         `json:"name"`
	ItemImage    *string        `json:"image_url"`
	Quantity     int32          `json:"quantity"`
	PriceAtOrder pgtype.Numeric `json:"price_at_order"`
}

type OrderResponse struct {
	ID         pgtype.UUID        `json:"id"`
	UserID     pgtype.UUID        `json:"user_id"`
	UserName   string             `json:"user_name"`
	OrderPrice pgtype.Numeric     `json:"order_price"`
	Status     string             `json:"status"`
	Items      []OrderItem        `json:"items"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz `json:"updated_at"`
}

func NewOrderService(repo OrderRepository, logger *slog.Logger) *OrderService {
	return &OrderService{
		repo:   repo,
		logger: logger,
	}
}

// GetOrderByID retrieves an order with all its items by ID
func (s *OrderService) GetOrderByID(ctx context.Context, id string) (*OrderResponse, error) {
	s.logger.Info("fetching order by id", "order_id", id)

	var uuid pgtype.UUID
	if err := uuid.Scan(id); err != nil {
		s.logger.Error("invalid order ID", "id", id, "error", err)
		return nil, fmt.Errorf("invalid order ID")
	}

	rows, err := s.repo.GetOrderByID(ctx, uuid)
	if err != nil {
		s.logger.Error("failed to get order from repository", "order_id", id, "error", err)
		return nil, err
	}

	if len(rows) == 0 {
		s.logger.Warn("order not found", "order_id", id)
		return nil, fmt.Errorf("order not found")
	}

	var orderResp *OrderResponse
	for _, row := range rows {
		if orderResp == nil {
			orderResp = &OrderResponse{
				ID:         row.ID,
				UserID:     row.UserID,
				UserName:   row.UserName,
				OrderPrice: row.OrderPrice,
				Status:     string(row.Status),
				Items:      []OrderItem{},
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			}
		}

		if row.ItemProductID.Valid {
			orderResp.Items = append(orderResp.Items, OrderItem{
				ID:           row.ItemID,
				ItemID:       row.ItemProductID,
				ItemCode:     *row.Code,
				ItemName:     *row.Name,
				ItemImage:    row.ImageUrl,
				Quantity:     *row.Quantity,
				PriceAtOrder: row.PriceAtOrder,
			})
		}
	}

	s.logger.Info("successfully fetched order",
		"order_id", orderResp.ID.String(),
		"user_name", orderResp.UserName,
		"items_count", len(orderResp.Items),
	)

	return orderResp, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status repository.OrderStatus) error {
	s.logger.Info("updating order status",
		"order_id", orderID,
		"status", status,
	)

	var uuid pgtype.UUID
	if err := uuid.Scan(orderID); err != nil {
		s.logger.Error("invalid order ID", "id", orderID, "error", err)
		return fmt.Errorf("invalid order ID")
	}

	_, err := s.repo.UpdateOrderStatus(ctx, repository.UpdateOrderStatusParams{
		ID:     uuid,
		Status: status,
	})
	if err != nil {
		s.logger.Error("failed to update order status", "order_id", orderID, "status", status, "error", err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	s.logger.Info("order status updated successfully",
		"order_id", orderID,
		"status", status,
	)

	return nil
}
