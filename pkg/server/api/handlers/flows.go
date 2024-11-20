package handlers

import (
	"encoding/json"
	"net/http"
)

type FlowHandler struct {
	flows FlowManager
}

type FlowManager interface {
	CreateFlow(id string, config map[string]interface{}) error
	GetFlow(id string) (map[string]interface{}, error)
}

func NewFlowHandler(fm FlowManager) *FlowHandler {
	return &FlowHandler{
		flows: fm,
	}
}

func (h *FlowHandler) CreateFlow(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID    string                 `json:"id"`
		Nodes map[string]interface{} `json:"nodes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.flows.CreateFlow(req.ID, req.Nodes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"id": req.ID,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
