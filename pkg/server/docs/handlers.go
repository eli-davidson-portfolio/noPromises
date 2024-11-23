package docs

import (
	"encoding/json"
	"net/http"
)

// HandleFlowDiagram serves a flow diagram
func (s *Server) HandleFlowDiagram(w http.ResponseWriter, _ *http.Request, id string) {
	diagram, err := s.mermaidGen.GenerateDiagram(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"diagram": diagram,
	}); err != nil {
		s.logDebug("Error encoding diagram response: %v", err)
	}
}
