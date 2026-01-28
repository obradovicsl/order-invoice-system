package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GetAllOrdersResponse struct {
	Orders []GetOrderResponse `json:"orders"`
}

type GetOrderResponse struct {
	ID         pgtype.UUID         `json:"id"`
	UserID     pgtype.UUID         `json:"user_id"`
	UserName   string              `json:"user_name"`
	Status     string              `json:"status"`
	OrderPrice float64             `json:"order_price"`
	Items      []OrderItemResponse `json:"items"`
}

func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("received get all orders request")

	orders, err := h.service.GetAllOrders(r.Context())
	if err != nil {
		h.logger.Error("failed to fetch orders", "error", err)
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	resp := GetAllOrdersResponse{
		Orders: make([]GetOrderResponse, len(orders)),
	}
	for i, order := range orders {
		resp.Orders[i] = GetOrderResponse{
			ID:         order.ID,
			UserID:     order.UserID,
			UserName:   order.UserName,
			Status:     order.Status,
			OrderPrice: numericToFloat(order.OrderPrice),
			Items:      make([]OrderItemResponse, len(order.Items)),
		}
		for j, item := range order.Items {
			resp.Orders[i].Items[j] = OrderItemResponse{
				ID:           item.ID,
				ItemID:       item.ItemID,
				ItemCode:     item.ItemCode,
				ItemName:     item.ItemName,
				ItemImage:    item.ItemImage,
				Quantity:     item.Quantity,
				PriceAtOrder: item.PriceAtOrder,
			}
		}
	}

	h.logger.Info("fetched all orders successfully",
		"total_orders", len(resp.Orders),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	h.logger.Info("received get order by id request", "order_id", idStr)

	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		h.logger.Error("invalid order ID", "id", idStr, "error", err)
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), idStr)
	if err != nil {
		h.logger.Error("failed to fetch order", "order_id", idStr, "error", err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	resp := GetOrderResponse{
		ID:         order.ID,
		UserID:     order.UserID,
		UserName:   order.UserName,
		Status:     order.Status,
		OrderPrice: numericToFloat(order.OrderPrice),
		Items:      make([]OrderItemResponse, len(order.Items)),
	}
	for i, item := range order.Items {
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

	h.logger.Info("fetched order successfully",
		"order_id", resp.ID.String(),
		"user_name", resp.UserName,
		"status", resp.Status,
		"items_count", len(resp.Items),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
