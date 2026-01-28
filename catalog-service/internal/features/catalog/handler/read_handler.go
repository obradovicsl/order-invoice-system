package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetAllCatalogItems je handler za GET /items
func (h *CatalogHandler) GetAllCatalogItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetAllCatalogItems(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch catalog items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(items)
}

// GetCatalogItemByID je handler za GET /items/{id}
func (h *CatalogHandler) GetCatalogItemByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	var id pgtype.UUID
	if err := id.Scan(idStr); err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.service.GetCatalogItemByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(item)
}
