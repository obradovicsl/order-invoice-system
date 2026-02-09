package handler

import (
	"encoding/json"
	"net/http"
)

type GetPresignedUploadURLRequest struct {
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

type GetPresignedUploadURLResponse struct {
	UploadURL   string `json:"upload_url"`
	DownloadURL string `json:"download_url"`
}

func (h *CatalogHandler) GetPresignedUploadURL(w http.ResponseWriter, r *http.Request) {
	var req GetPresignedUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FileName == "" || req.FileType == "" {
		h.logger.Error("missing required fields", "file_name", req.FileName, "file_type", req.FileType)
		http.Error(w, "file_name and file_type are required", http.StatusBadRequest)
		return
	}

	h.logger.Info("received presigned upload URL request",
		"file_name", req.FileName,
		"file_type", req.FileType,
	)

	uploadURL, downloadURL, err := h.service.GetPresignedUploadURL(r.Context(), req.FileName, req.FileType)
	if err != nil {
		h.logger.Error("failed to generate presigned upload URL",
			"file_name", req.FileName,
			"error", err,
		)
		http.Error(w, "Failed to generate presigned upload URL", http.StatusInternalServerError)
		return
	}

	resp := GetPresignedUploadURLResponse{
		UploadURL:   uploadURL,
		DownloadURL: downloadURL,
	}

	h.logger.Info("presigned URLs generated successfully", "file_name", req.FileName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
