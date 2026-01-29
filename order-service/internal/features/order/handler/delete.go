package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *OrderHandler) DeleteOrderById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.logger.Info("received delete order request", "order_id", id)

	if err := h.service.DeleteOrderById(r.Context(), id); err != nil {
		h.logger.Error("failed to delete order", "error", err, "order_id", id)
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}

	h.logger.Info("deleted order successfully", "order_id", id)
	w.WriteHeader(http.StatusNoContent)
}
