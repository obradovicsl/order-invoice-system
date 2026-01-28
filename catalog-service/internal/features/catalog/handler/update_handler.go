package handler

import (
	"catalog-service/internal/repository"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UpdateItemQuantityRequest predstavlja request za ažuriranje količine
type UpdateItemQuantityRequest struct {
	StockQuantity int32 `json:"stock_quantity"`
}

// UpdateItemQuantity je handler za PUT /items/{id}/quantity
func (h *CatalogHandler) UpdateItemQuantity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req UpdateItemQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := repository.UpdateItemQuantityParams{
		ID:            id,
		StockQuantity: req.StockQuantity,
	}

	item, err := h.service.UpdateItemQuantity(r.Context(), params)
	if err != nil {
		http.Error(w, "Failed to update item quantity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}
