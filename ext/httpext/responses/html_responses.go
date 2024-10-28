package responses

import (
	"log/slog"
	"net/http"
)

// Html writes the provided HTML content to the client, using the given HTTP status code.
// It sets the Content-Type header to "text/html".
// If there's an error during writing the response, it logs the error and returns a 500 Internal Server Error.
func Html(w http.ResponseWriter, htmlContent string, statusCode int) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(htmlContent))
	if err != nil {
		slog.Error("Failed to write HTML response to client", "error", err)
		return err
	}
	return nil
}

// HtmlOK writes the provided HTML content to the client with a 200 OK status code.
// If there's an error during the response process, it logs the error and returns a 500 Internal Server Error.
func HtmlOK(w http.ResponseWriter, htmlContent string) {
	err := Html(w, htmlContent, http.StatusOK)
	if err != nil {
		slog.Error("Failed to return HTML content", "error", err)
		http.Error(w, "<h1>Internal Server Error</h1>", http.StatusInternalServerError)
		return
	}
}

// HtmlNotFound writes an HTML response to the client with a 404 Not Found status code.
// It typically indicates that the requested page could not be found.
// If there's an error during the response process, it logs the error and returns a 500 Internal Server Error.
func HtmlNotFound(w http.ResponseWriter, htmlContent string) {
	err := Html(w, htmlContent, http.StatusNotFound)
	if err != nil {
		slog.Error("Failed to return not found message as HTML", "error", err)
		http.Error(w, "<h1>Internal Server Error</h1>", http.StatusInternalServerError)
	}
}
