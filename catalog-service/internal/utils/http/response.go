package http

import (
	"encoding/json"
	"net/http"
)

// ErrorResponseBody represents the structure of an error response
type ErrorResponseBody struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteErrorResponse writes a JSON error response with code, message, and internal code
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string, internalCode string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponseBody{
		Error:   message,
		Code:    internalCode,
		Message: message,
	}

	err := json.NewEncoder(w).Encode(response)

	return err
}

// WriteJSONResponse writes a JSON response with the given status code and data
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)

	return err
}
