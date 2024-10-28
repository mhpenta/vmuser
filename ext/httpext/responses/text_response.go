package responses

import (
	"log/slog"
	"net/http"
)

// Text writes the provided plain text content to the client, using the given HTTP status code.
// It sets the Content-Type header to "text/plain".
// If there's an error during writing the response, it logs the error and returns it.
func Text(w http.ResponseWriter, textContent string, statusCode int) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	_, err := w.Write([]byte(textContent))
	if err != nil {
		slog.Error("Failed to write text response to client", "error", err)
		return err
	}
	return nil
}

// TextOK writes the provided plain text content to the client with a 200 OK status code.
// If there's an error during the response process, it logs the error and returns a 500 Internal Server Error.
func TextOK(w http.ResponseWriter, textContent string) {
	err := Text(w, textContent, http.StatusOK)
	if err != nil {
		slog.Error("Failed to return text content", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
