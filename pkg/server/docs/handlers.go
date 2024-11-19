package docs

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// handleNetworkDiagram generates and serves a Mermaid diagram for a network
func (s *Server) handleNetworkDiagram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	networkID := vars["id"]

	diagram, err := s.mermaidGen.GenerateFlowDiagram(networkID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{
		"diagram": diagram,
	}); err != nil {
		log.Printf("Error encoding diagram response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleLiveDiagram handles WebSocket connections for live diagram updates
func (s *Server) handleLiveDiagram(w http.ResponseWriter, _ *http.Request) {
	// For now, just return switching protocols status
	// WebSocket implementation will be added later
	w.WriteHeader(http.StatusSwitchingProtocols)
}

// handleSwaggerUI serves the Swagger UI for API documentation
func (s *Server) handleSwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(`<!DOCTYPE html>
<html>
  <head>
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-16x16.png" sizes="16x16" />
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function() {
        const ui = SwaggerUIBundle({
          url: "/api/swagger.json",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
        });
        window.ui = ui;
      }
    </script>
  </body>
</html>`))
	if err != nil {
		log.Printf("Error writing Swagger UI response: %v", err)
	}
}
