package handler

import (
	"catalog-service/internal/repository"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCatalogItemRequest struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	StockQuantity int32   `json:"stock_quantity"`
	ImageUrl      *string `json:"image_url,omitempty"`
}

func (h *CatalogHandler) CreateCatalogItem(w http.ResponseWriter, r *http.Request) {
	var req CreateCatalogItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	priceDecimal := new(big.Float)
	priceDecimal.SetString(fmt.Sprintf("%.2f", req.Price))
	intValue := new(big.Int)
	priceDecimal.Mul(priceDecimal, big.NewFloat(100)).Int(intValue)

	price := pgtype.Numeric{
		Int:   intValue,
		Exp:   -2,
		Valid: true,
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
