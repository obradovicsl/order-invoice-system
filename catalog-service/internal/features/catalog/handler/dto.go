package handler

import (
	"catalog-service/internal/repository"
	"fmt"
	"time"
)

// ItemResponse predstavlja DTO response za stavku kataloga
type ItemResponse struct {
	ID            string  `json:"id"`
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Price         string  `json:"price"`
	StockQuantity int32   `json:"stock_quantity"`
	ImageUrl      *string `json:"image_url,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ItemsResponse predstavlja DTO response za listu stavki
type ItemsResponse struct {
	Data      []ItemResponse `json:"data"`
	Count     int            `json:"count"`
	Timestamp string         `json:"timestamp"`
}

// MapItemToResponse mapira repository Item na ItemResponse
func MapItemToResponse(item repository.Item) ItemResponse {
	price := ""
	if item.Price.Valid {
		val, err := item.Price.Float64Value()
		if err == nil {
			price = fmt.Sprintf("%.2f", val.Float64)
		}
	}

	return ItemResponse{
		ID:            item.ID.String(),
		Code:          item.Code,
		Name:          item.Name,
		Price:         price,
		StockQuantity: item.StockQuantity,
		ImageUrl:      item.ImageUrl,
		CreatedAt:     item.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:     item.UpdatedAt.Time.Format(time.RFC3339),
	}
}

// MapItemsToResponse mapira listu repository Item na ItemsResponse
func MapItemsToResponse(items []repository.Item) ItemsResponse {
	responses := make([]ItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, MapItemToResponse(item))
	}

	return ItemsResponse{
		Data:      responses,
		Count:     len(responses),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}
