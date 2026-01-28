package handler

import (
	"catalog-service/internal/repository"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

// CreateCatalogItemRequest predstavlja request za kreiranje stavke
type CreateCatalogItemRequest struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Price         string  `json:"price"`
	StockQuantity int32   `json:"stock_quantity"`
	ImageUrl      *string `json:"image_url,omitempty"`
}

// CreateCatalogItem je handler za POST /items
func (h *CatalogHandler) CreateCatalogItem(w http.ResponseWriter, r *http.Request) {
	var req CreateCatalogItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Konverzija string cene u pgtype.Numeric
	var price pgtype.Numeric
	if err := price.Scan(req.Price); err != nil {
		h.logger.Error("invalid price format", "price", req.Price, "error", err)
		http.Error(w, "Invalid price format", http.StatusBadRequest)
		return
	}

	params := repository.CreateCatalogItemParams{
		Code:          req.Code,
		Name:          req.Name,
		Price:         price,
		StockQuantity: req.StockQuantity,
		ImageUrl:      req.ImageUrl,
	}

	item, err := h.service.CreateCatalogItem(r.Context(), params)
	if err != nil {
		h.logger.Error("failed to create catalog item", "error", err)
		http.Error(w, "Failed to create catalog item", http.StatusInternalServerError)
		return
	}

	response := MapItemToResponse(item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
