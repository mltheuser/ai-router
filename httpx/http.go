package httpx

import (
	"encoding/json"
	"net/http"
)

// DecodeJSON decodes a JSON request body into v. A malformed body becomes a
// 400 *Error carrying the decoder's explanation.
func DecodeJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return NewError(http.StatusBadRequest, "invalid request body: "+err.Error())
	}
	return nil
}

// WriteJSON writes v as a 200 JSON response.
func WriteJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
