package handler

import (
	"encoding/json"
	"net/http"
	svc "order-service/internal/features/order"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateOrderItemRequest struct {
	ItemID   pgtype.UUID `json:"item_id"`
	Quantity int32       `json:"quantity"`
}

type CreateOrderRequest struct {
	UserName string                   `json:"user_name"`
	Items    []CreateOrderItemRequest `json:"items"`
}

type OrderItemResponse struct {
	ID           pgtype.UUID    `json:"id"`
	ItemID       pgtype.UUID    `json:"item_id"`
	ItemCode     string         `json:"code"`
	ItemName     string         `json:"name"`
	ItemImage    *string        `json:"image_url"`
	Quantity     int32          `json:"quantity"`
	PriceAtOrder pgtype.Numeric `json:"price_at_order"`
}

type CreateOrderResponse struct {
	ID         pgtype.UUID         `json:"id"`
	UserID     pgtype.UUID         `json:"user_id"`
	UserName   string              `json:"user_name"`
	Status     string              `json:"status"`
	OrderPrice float64             `json:"order_price"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  pgtype.Timestamptz  `json:"created_at"`
	UpdatedAt  pgtype.Timestamptz  `json:"updated_at"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("received create order request",
		"user_name", req.UserName,
		"items_count", len(req.Items),
	)

	userID, _ := uuid.Parse("00000000-0000-0000-0000-000000000001")
	serviceReq := svc.CreateOrderRequest{
		UserID:   pgtype.UUID{Bytes: userID, Valid: true},
		UserName: req.UserName,
		Items:    make([]svc.CreateOrderItemRequest, len(req.Items)),
	}

	for i, item := range req.Items {
		serviceReq.Items[i] = svc.CreateOrderItemRequest{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
		}
	}

	orderResp, err := h.service.CreateOrder(r.Context(), serviceReq)
	if err != nil {
		h.logger.Error("failed to create order", "error", err, "user_name", req.UserName)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CreateOrderResponse{
		ID:         orderResp.ID,
		UserID:     orderResp.UserID,
		UserName:   orderResp.UserName,
		Status:     orderResp.Status,
		OrderPrice: numericToFloat(orderResp.OrderPrice),
		Items:      make([]OrderItemResponse, len(orderResp.Items)),
		CreatedAt:  orderResp.CreatedAt,
		UpdatedAt:  orderResp.UpdatedAt,
	}
	for i, item := range orderResp.Items {
		resp.Items[i] = OrderItemResponse{
			ID:           item.ID,
			ItemID:       item.ItemID,
			ItemCode:     item.ItemCode,
			ItemName:     item.ItemName,
			ItemImage:    item.ItemImage,
			Quantity:     item.Quantity,
			PriceAtOrder: item.PriceAtOrder,
		}
	}

	h.logger.Info("order created successfully",
		"order_id", resp.ID.String(),
		"user_name", resp.UserName,
		"status", resp.Status,
		"items_count", len(resp.Items),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
