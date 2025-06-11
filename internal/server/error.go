package server

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

// writeErrorResponse writes a JSON error response to the provided http.ResponseWriter.
func writeErrorResponse(w http.ResponseWriter, statusCode int, err string, message string) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	response := map[string]string{
		"error":   err,
		"message": message,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		glog.Error("Failed to encode error response:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)

		return
	}
}
