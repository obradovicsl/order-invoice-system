package http

import (
	"encoding/json"
	"net/http"
)

// DecodeJSONRequest decodes a JSON request body into the provided struct
func DecodeJSONRequest[T any](r *http.Request, dst *T) error {
	err := json.NewDecoder(r.Body).Decode(dst)

	return err
}
