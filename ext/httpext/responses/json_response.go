package responses

import (
	//"github.com/goccy/go-json"
	"encoding/json"
	"log/slog"
	"net/http"
)

// JsonEncodePrefix defines the prefix to use when marshalling JSON.
const JsonEncodePrefix = ""

// JsonEncodeIndent defines the indentation to use when marshalling JSON.
const JsonEncodeIndent = "  "

// Json writes the provided object as a JSON response to the client, using the given HTTP status code.
// It sets the Content-Type header to "application/json".
// If there's an error during marshalling or writing the response, it logs the error and returns a 500 Internal Server Error.
func Json(w http.ResponseWriter, obj interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonOutput, err := json.MarshalIndent(obj, JsonEncodePrefix, JsonEncodeIndent)
	if err != nil {
		slog.Error("Error marshalling object to JSON", "error", err)
		return err
	}
	_, err = w.Write(jsonOutput)
	if err != nil {
		slog.Error("Failed to write JSON response to client", "error", err)
		return err
	}
	return nil
}

// JsonOK writes the provided object as a JSON response to the client with a 200 OK status code.
// If there's an error during the response process, it logs the error and returns a 500 Internal Server Error.
func JsonOK(w http.ResponseWriter, obj interface{}) {
	err := Json(w, obj, http.StatusOK)
	if err != nil {
		slog.Error("Failed to return object as JSON", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// JsonOKFromString writes the provided string as a JSON response to the client with a 200 OK status code.
// If there's an error during the response process, it logs the error and returns a 500 Internal Server Error.
func JsonOKFromString(w http.ResponseWriter, jsonString []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(jsonString)
	if err != nil {
		slog.Error("Failed to write JSON response to client", "error", err)
		return
	}
	return
}

// JsonDataNotFound writes a JSON response to the client with a 404 Not Found status code.
// It typically indicates that the requested data could not be found.
// If there's an error during the response process, it logs the error and returns a 500 Internal Server Error.
func JsonDataNotFound(w http.ResponseWriter, message string) {
	responseObj := map[string]string{"error": message}
	err := Json(w, responseObj, http.StatusNotFound)
	if err != nil {
		slog.Error("Failed to return not found message as JSON", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// JsonReturnJson writes the provided object as a JSON response to the client, using the given HTTP status code.
// It sets the Content-Type header to "application/json".
// If there's an error during marshalling or writing the response, it logs the error and returns a 500 Internal Server Error.
// Function returns Json written to writer.
func JsonReturnJson(w http.ResponseWriter, obj interface{}, statusCode int) ([]byte, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonOutput, err := json.MarshalIndent(obj, JsonEncodePrefix, JsonEncodeIndent)
	if err != nil {
		slog.Error("Error marshalling object to JSON", "error", err)
		return []byte{}, err
	}
	_, err = w.Write(jsonOutput)
	if err != nil {
		slog.Error("Failed to write JSON response to client", "error", err)
		return []byte{}, err
	}
	return jsonOutput, nil
}

// JsonOKReturnJson writes the provided object as a JSON response to the client with a 200 OK status code and returns
// the json byte array to the caller.
func JsonOKReturnJson(w http.ResponseWriter, obj interface{}) []byte {
	jsonOutput, err := JsonReturnJson(w, obj, http.StatusOK)
	if err != nil {
		slog.Error("Failed to return object as JSON", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return []byte{}
	}
	return jsonOutput
}

// WriteJsonBytes writes the JSON []byte to the client, using the given HTTP status code.
// It sets the Content-Type header to "application/json".
// If there's an error during marshalling or writing the response, it logs the error and returns a 500 Internal Server Error.
func WriteJsonBytes(w http.ResponseWriter, jsonOutput []byte, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err := w.Write(jsonOutput)
	if err != nil {
		slog.Error("Failed to write JSON response to client", "error", err)
		return err
	}
	return nil
}

// JsonError writes an error message as a JSON response to the client, using the given HTTP status code.
// It sets the Content-Type header to "application/json".
// If there's an error during marshalling or writing the response, it logs the error and returns a 500 Internal Server Error.
func JsonError(w http.ResponseWriter, serverError int, errorMessage string) {
	// Creating a map to hold the error message
	responseObj := map[string]string{"error": errorMessage}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(serverError)
	jsonOutput, err := json.MarshalIndent(responseObj, JsonEncodePrefix, JsonEncodeIndent)
	if err != nil {
		slog.Error("Error marshalling error message to JSON", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonOutput)
	if err != nil {
		slog.Error("Failed to write JSON error response to client", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
