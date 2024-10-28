package responses

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(ErrorResponse{
		Code:    statusCode,
		Message: message,
		Details: details,
	})
	if err != nil {
		slog.Error("Error encoding JSON when marshalling error json", "error", err)
		return
	}
}
