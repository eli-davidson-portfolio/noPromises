package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// ProcessHandler handles process-related requests
type ProcessHandler struct {
	// Add required dependencies
}

// ListProcesses handles GET /processes
func (h *ProcessHandler) ListProcesses(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": "not implemented",
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
