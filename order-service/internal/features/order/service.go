package order

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"order-service/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type Repository interface {
	CreateOrder(ctx context.Context, arg repository.CreateOrderParams) (repository.Order, error)
	CreateOrderItem(ctx context.Context, arg repository.CreateOrderItemParams) (repository.OrderItem, error)
	GetAllOrders(ctx context.Context) ([]repository.GetAllOrdersRow, error)
	GetItemByID(ctx context.Context, id pgtype.UUID) (repository.GetItemByIDRow, error)
	GetOrderByID(ctx context.Context, id pgtype.UUID) ([]repository.GetOrderByIDRow, error)
	UpdateItemStock(ctx context.Context, arg repository.UpdateItemStockParams) (repository.Item, error)

	// Order
	UpdateOrderStatus(ctx context.Context, arg repository.UpdateOrderStatusParams) (repository.Order, error)
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

type CreateOrderItemRequest struct {
	ItemID   pgtype.UUID `json:"item_id"`
	Quantity int32       `json:"quantity"`
}

type CreateOrderRequest struct {
	UserID   pgtype.UUID              `json:"user_id"`
	UserName string                   `json:"user_name"`
	Items    []CreateOrderItemRequest `json:"items"`
}

func (s *Service) GetAllOrders(ctx context.Context) ([]OrderResponse, error) {
	s.logger.Info("fetching all orders from repository")

	rows, err := s.repo.GetAllOrders(ctx)
	if err != nil {
		s.logger.Error("failed to get all orders from repository", "error", err)
		return nil, err
	}

	ordersMap := make(map[string]*OrderResponse)

	for _, row := range rows {
		orderID := row.ID.String()

		if _, exists := ordersMap[orderID]; !exists {
			ordersMap[orderID] = &OrderResponse{
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
			ordersMap[orderID].Items = append(ordersMap[orderID].Items, OrderItem{
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

	orders := make([]OrderResponse, 0, len(ordersMap))
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	s.logger.Info("successfully fetched all orders",
		"total_orders", len(orders),
		"total_rows", len(rows),
	)

	return orders, nil
}

func (s *Service) GetOrderByID(ctx context.Context, id string) (*OrderResponse, error) {
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

func (s *Service) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	s.logger.Info("starting create order process",
		"user_id", req.UserID.String(),
		"user_name", req.UserName,
		"items_count", len(req.Items),
	)

	if len(req.Items) == 0 {
		s.logger.Error("order must have at least one item", "user_name", req.UserName)
		return nil, fmt.Errorf("order must have at least one item")
	}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			s.logger.Error("invalid quantity", "item_id", item.ItemID.String(), "quantity", item.Quantity)
			return nil, fmt.Errorf("quantity must be greater than 0")
		}
	}

	s.logger.Info("validating items and stock")

	totalPrice := float64(0)

	// Validate all items and calculate total price
	for _, item := range req.Items {
		catalogItem, err := s.repo.GetItemByID(ctx, item.ItemID)
		if err != nil {
			s.logger.Error("item not found", "item_id", item.ItemID.String(), "error", err)
			return nil, fmt.Errorf("item %s not found", item.ItemID)
		}

		if catalogItem.StockQuantity < item.Quantity {
			s.logger.Warn("insufficient stock",
				"item_id", item.ItemID.String(),
				"item_code", catalogItem.Code,
				"available_stock", catalogItem.StockQuantity,
				"requested_quantity", item.Quantity,
			)
			return nil, fmt.Errorf("insufficient stock for item %s: have %d, need %d",
				catalogItem.Code, catalogItem.StockQuantity, item.Quantity)
		}

		s.logger.Info("item validated successfully",
			"item_id", item.ItemID.String(),
			"item_code", catalogItem.Code,
			"quantity", item.Quantity,
		)

		// Calculate total price: price * quantity
		priceFloat := catalogItem.Price.Int.Int64() * int64(item.Quantity) / int64(big.NewInt(10).Exp(big.NewInt(10), big.NewInt(int64(-catalogItem.Price.Exp)), nil).Int64())
		totalPrice += float64(priceFloat)
	}

	s.logger.Info("creating order in database", "user_name", req.UserName, "total_price", fmt.Sprintf("%.2f", totalPrice))

	// Convert totalPrice to pgtype.Numeric
	priceDecimal := new(big.Float)
	priceDecimal.SetString(fmt.Sprintf("%.2f", totalPrice))
	intValue := new(big.Int)
	priceDecimal.Mul(priceDecimal, big.NewFloat(100)).Int(intValue)
	orderPrice := pgtype.Numeric{
		Int:   intValue,
		Exp:   -2,
		Valid: true,
	}

	order, err := s.repo.CreateOrder(ctx, repository.CreateOrderParams{
		UserID:     req.UserID,
		UserName:   req.UserName,
		Status:     repository.OrderStatusPENDING,
		OrderPrice: orderPrice,
	})
	if err != nil {
		s.logger.Error("failed to create order in database", "error", err, "user_name", req.UserName)
		return nil, err
	}

	s.logger.Info("order created", "order_id", order.ID.String(), "status", order.Status)

	orderItems := make([]OrderItem, 0)

	// Update stock and create order items
	for _, item := range req.Items {
		catalogItem, _ := s.repo.GetItemByID(ctx, item.ItemID)

		s.logger.Info("updating item stock",
			"item_id", item.ItemID.String(),
			"quantity", item.Quantity,
		)

		_, err := s.repo.UpdateItemStock(ctx, repository.UpdateItemStockParams{
			ID:            item.ItemID,
			StockQuantity: item.Quantity,
		})
		if err != nil {
			s.logger.Error("failed to update stock", "item_id", item.ItemID.String(), "error", err)
			return nil, fmt.Errorf("failed to update stock for item")
		}

		s.logger.Info("stock updated successfully", "item_id", item.ItemID.String())

		s.logger.Info("creating order item",
			"order_id", order.ID.String(),
			"item_id", item.ItemID.String(),
		)

		orderItem, err := s.repo.CreateOrderItem(ctx, repository.CreateOrderItemParams{
			OrderID:      order.ID,
			ItemID:       item.ItemID,
			Quantity:     item.Quantity,
			PriceAtOrder: catalogItem.Price,
		})
		if err != nil {
			s.logger.Error("failed to create order item",
				"order_id", order.ID.String(),
				"item_id", item.ItemID.String(),
				"error", err,
			)
			return nil, err
		}

		s.logger.Info("order item created successfully",
			"order_item_id", orderItem.ID.String(),
			"item_code", catalogItem.Code,
		)

		orderItems = append(orderItems, OrderItem{
			ID:           orderItem.ID,
			ItemID:       orderItem.ItemID,
			ItemCode:     catalogItem.Code,
			ItemName:     catalogItem.Name,
			ItemImage:    catalogItem.ImageUrl,
			Quantity:     orderItem.Quantity,
			PriceAtOrder: orderItem.PriceAtOrder,
		})
	}

	s.logger.Info("order completed successfully",
		"order_id", order.ID.String(),
		"user_name", order.UserName,
		"total_items", len(orderItems),
		"status", order.Status,
	)

	return &OrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		UserName:   order.UserName,
		Status:     string(order.Status),
		OrderPrice: order.OrderPrice,
		Items:      orderItems,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}, nil
}
